package api

import (
	"encoding/json"
	util "github.com/devtron-labs/central-api/client"
	"github.com/devtron-labs/central-api/common"
	"github.com/devtron-labs/central-api/pkg"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strconv"
)

type RestHandler interface {
	GetReleases(w http.ResponseWriter, r *http.Request)
	ReleaseWebhookHandler(w http.ResponseWriter, r *http.Request)
}

func NewRestHandlerImpl(logger *zap.SugaredLogger, releaseNoteService pkg.ReleaseNoteService,
	webhookSecretValidator pkg.WebhookSecretValidator, client *util.GitHubClient) *RestHandlerImpl {
	return &RestHandlerImpl{
		logger:                 logger,
		releaseNoteService:     releaseNoteService,
		webhookSecretValidator: webhookSecretValidator,
		client:                 client,
	}
}

type RestHandlerImpl struct {
	logger                 *zap.SugaredLogger
	releaseNoteService     pkg.ReleaseNoteService
	webhookSecretValidator pkg.WebhookSecretValidator
	client                 *util.GitHubClient
}

func (impl RestHandlerImpl) WriteJsonResp(w http.ResponseWriter, err error, respBody interface{}, status int) {
	response := common.Response{}
	response.Code = status
	response.Status = http.StatusText(status)
	if err == nil {
		response.Result = respBody
	} else {
		apiErr := &common.ApiError{}
		apiErr.Code = "000" // 000=unknown
		apiErr.InternalMessage = err.Error()
		apiErr.UserMessage = respBody
		response.Errors = []*common.ApiError{apiErr}

	}
	b, err := json.Marshal(response)
	if err != nil {
		impl.logger.Error("error in marshaling err object", err)
		status = 500
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(b)
}

func (impl *RestHandlerImpl) GetReleases(w http.ResponseWriter, r *http.Request) {
	impl.logger.Info("get releases ..")
	response, err := impl.releaseNoteService.GetReleases()
	if err != nil {
		impl.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	impl.WriteJsonResp(w, nil, response, http.StatusOK)
	return
}

func (impl *RestHandlerImpl) ReleaseWebhookHandler(w http.ResponseWriter, r *http.Request) {
	impl.logger.Info("release webhook handler ..")

	// get git host Id and secret from request
	vars := mux.Vars(r)
	gitHostId, err := strconv.Atoi(vars["gitHostId"])
	secretFromRequest := vars["secret"]
	if err != nil {
		impl.logger.Errorw("Error in getting git host Id from request", "err", err)
		impl.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	impl.logger.Debug("gitHostId", gitHostId)
	impl.logger.Debug("secretFromRequest", secretFromRequest)

	// validate signature
	requestBodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		impl.logger.Errorw("Cannot read the request body:", "err", err)
		impl.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	isValidSig := impl.webhookSecretValidator.ValidateSecret(r, requestBodyBytes)
	impl.logger.Debug("Secret validation result ", isValidSig)
	if !isValidSig {
		impl.logger.Error("Signature mismatch")
		impl.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	// validate event type
	eventType := r.Header.Get(impl.client.GitHubConfig.GitHubEventTypeHeader)
	impl.logger.Debugw("eventType : ", eventType)
	if len(eventType) == 0 && eventType != "release" {
		impl.logger.Errorw("Event type not known ", eventType)
		impl.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	flag, err := impl.releaseNoteService.UpdateReleases(requestBodyBytes)
	if err != nil {
		impl.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	impl.WriteJsonResp(w, err, flag, http.StatusOK)
	return
}