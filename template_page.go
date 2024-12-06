package sitepages

import (
	"html/template"
	"log"
	"os"
	"strings"
)

func LoadAllTemplatePages(frontFolder string, templateFolder string) map[string]*template.Template {
	retval := make(map[string]*template.Template)
	funcMap := template.FuncMap{
		"split": split,
		"gentoken": func(data TemplateData, purpose string) string {
			salt := GenerateSalt(data.Nonce, purpose)
			return GenerateTokenFromSalt(data.SessionToken, salt)
		},
	}

	pagefiles, err := os.ReadDir(frontFolder)
	if err != nil {
		log.Fatal("Error reading directory:", err)
	}

	for _, filename := range pagefiles {
		tmpl, err := template.New(filename.Name()).Funcs(funcMap).ParseFiles(frontFolder + filename.Name())
		if err != nil {
			log.Fatal(err)
		}

		tmpl, err = tmpl.ParseGlob(templateFolder + "*.html")
		if err != nil {
			log.Fatal(err)
		}

		retval[filename.Name()] = tmpl

	}
	return retval
}

func split(s, sep string) []string {
	return strings.Split(s, sep)
}
