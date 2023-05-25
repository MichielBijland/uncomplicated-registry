package module

import (
	"github.com/gofiber/fiber/v2"
)

type listRequest struct {
	namespace string
	name      string
	provider  string
}

type listResponseVersion struct {
	Version string `json:"version,omitempty"`
}

type listResponseModule struct {
	Versions []listResponseVersion `json:"versions,omitempty"`
}

type listResponse struct {
	Modules []listResponseModule `json:"modules,omitempty"`
}

func listEndpoint(svc Service) fiber.Handler {
	return func(c *fiber.Ctx) error {

		res, err := svc.ListModuleVersions(c.Context(), c.Params("namespace"), c.Params("name"), c.Params("provider"))
		if err != nil {
			return errorHandler(c, err)
		}

		if len(res) == 0 {
			return notFoundHandler(c)
		}

		var versions []listResponseVersion

		for _, module := range res {
			versions = append(versions, listResponseVersion{
				Version: module.Version,
			})
		}

		return c.JSON(listResponse{
			Modules: []listResponseModule{
				{
					Versions: versions,
				},
			},
		})
	}
}

func downloadEndpoint(svc Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		res, err := svc.GetModule(c.Context(), c.Params("namespace"), c.Params("name"), c.Params("provider"), c.Params("version"))
		if err != nil {
			return errorHandler(c, err)
		}

		c.Set("X-Terraform-Get", res.DownloadURL)
		return c.SendStatus(fiber.StatusNoContent)
	}
}
