package swagger

import (
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"fmt"
)

var (
	bundle *i18n.Bundle
	localesDir string
	langTemplateData map[string]map[string]string
)

func SetLocalesDir(dir string)  {
	bundle = &i18n.Bundle{DefaultLanguage: language.English}
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	localesDir = dir
	langTemplateData = make(map[string]map[string]string)

	loadLocale("en-us")
}

func loadLocale(lang string) error {
	messageFile, err := bundle.LoadMessageFile(localesDir + "/" + lang + ".toml")
	if err != nil {
		fmt.Println(err)
		return err
	}

	templateData := make(map[string]string)
	for _, v := range messageFile.Messages {
		templateData[v.ID] = v.Other
	}

	langTemplateData[lang] = templateData
	return nil
}

func TranslateText(text, lang string) string {
	if bundle == nil {
		return text
	}
	localizer := i18n.NewLocalizer(bundle, lang)

	templateData, ok := langTemplateData[lang]
	if !ok {
		if err := loadLocale(lang); err == nil {
			templateData = langTemplateData[lang]
		} else {
			templateData = langTemplateData["en-us"]
		}
	}

	if templateData == nil{
		return text
	}

	return localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:"ApiDoc",
			Other:text,
		},
		TemplateData: templateData,
	})
}

func test()  {
	text := "we {{.SayHello}}, ok {{.Success}}"

	fmt.Println(TranslateText(text, "en-us"))
	fmt.Println(TranslateText(text, "zh-cn"))
}