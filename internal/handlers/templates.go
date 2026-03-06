package handlers

import (
	"embed"
	"net/http"
	"text/template"

	"github.com/amieldelatorre/shurl/internal/config"
	"github.com/amieldelatorre/shurl/internal/types"
	"github.com/amieldelatorre/shurl/internal/utils"
)

//go:embed templates
var templateFS embed.FS

var templates = template.Must(template.ParseFS(templateFS, "templates/*"))

type TemplateHandler struct {
	Logger  utils.CustomJsonLogger
	Config  *config.Config
	BaseUrl string
}

func NewTemplateHandler(logger utils.CustomJsonLogger, baseUrl string, config *config.Config) TemplateHandler {
	return TemplateHandler{Logger: logger, BaseUrl: baseUrl, Config: config}
}

func (h *TemplateHandler) GetIndexJs(w http.ResponseWriter, r *http.Request) {
	templateData := map[string]any{
		"apiUrl":            h.BaseUrl,
		"allowRegistration": h.Config.Server.AllowRegistration,
		"allowLogin":        h.Config.Server.AllowLogin,
		"allowAnonymous":    h.Config.Server.AllowAnonymous,
	}

	w.Header().Set("Content-Type", "text/javascript")
	err := templates.ExecuteTemplate(w, "shared.js", templateData)
	if err != nil {
		EncodeResponse[types.ErrorResponse](h.Logger, r.Context(), w, http.StatusInternalServerError, types.ErrorResponse{Errors: []string{"Something is wrong with the server. Please try again later"}})
		h.Logger.Error(r.Context(), err.Error())
		return
	}
}
