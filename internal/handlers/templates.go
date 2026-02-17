package handlers

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/amieldelatorre/shurl/internal/utils"
)

//go:embed templates
var templateFS embed.FS

var templates = template.Must(template.ParseFS(templateFS, "templates/*"))

type TemplateHandler struct {
	Logger  utils.CustomJsonLogger
	BaseUrl string
}

func NewTemplateHandler(logger utils.CustomJsonLogger, baseUrl string) TemplateHandler {
	return TemplateHandler{Logger: logger, BaseUrl: baseUrl}
}

func (h *TemplateHandler) GetIndexJs(w http.ResponseWriter, r *http.Request) {
	templateData := map[string]string{"apiUrl": h.BaseUrl}

	w.Header().Set("Content-Type", "text/javascript")
	err := templates.ExecuteTemplate(w, "index.js", templateData)
	if err != nil {
		EncodeResponse[types.ErrorResponse](w, http.StatusInternalServerError, types.ErrorResponse{Error: "Something is wrong with the server. Please try again later"})
		h.Logger.Error(r.Context(), err.Error())
		return
	}
}
