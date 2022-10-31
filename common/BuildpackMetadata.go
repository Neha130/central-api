package common

type Language string

const (
	NODE   Language = "Node"
	JAVA   Language = "Java"
	PYTHON Language = "Python"
	PHP    Language = "PHP"
	RUBY   Language = "Ruby"
	GO     Language = "Go"
	DOTNET Language = ".NET"
)

type BuildPackMetadata struct {
	Builders        []*Builder
	LanguageBuilder []*LanguageBuilder
}

type Builder struct {
	Id              string
	ConfigLink      string
	EntryPointParam string
	LanguageSupport []*LanguageSupport
}

type LanguageSupport struct {
	Language            Language
	BuilderLangEnvParam string
	Versions            []string
}

type BuilderLanguageMetadata struct {
	Id                  string
	BuilderLangEnvParam string
}

type LanguageBuilder struct {
	Language                Language
	LanguageIcon            string
	Versions                []string
	BuilderLanguageMetadata []*BuilderLanguageMetadata
}
