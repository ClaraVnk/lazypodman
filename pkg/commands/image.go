package commands

import (
	"context"
	"strings"

	"github.com/ClaraVnk/lazypodman/pkg/domain"
	"github.com/ClaraVnk/lazypodman/pkg/runtime"
	"github.com/ClaraVnk/lazypodman/pkg/utils"
	"github.com/fatih/color"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
)

// Image : an OCI image known to the runtime.
type Image struct {
	Name          string
	Tag           string
	ID            string
	Image         domain.ImageInfo
	OSCommand     *OSCommand
	Log           *logrus.Entry
	DockerCommand LimitedDockerCommand
	Runtime       runtime.ContainerRuntime
}

// Remove removes the image.
func (i *Image) Remove(options runtime.RemoveImageOptions) error {
	return i.Runtime.RemoveImage(context.Background(), i.ID, options)
}

func getHistoryResponseItemDisplayStrings(layer domain.ImageHistoryItem) []string {
	tag := ""
	if len(layer.Tags) > 0 {
		tag = layer.Tags[0]
	}

	id := strings.TrimPrefix(layer.ID, "sha256:")
	if len(id) > 10 {
		id = id[0:10]
	}
	idColor := color.FgWhite
	if id == "<missing>" {
		idColor = color.FgBlue
	}

	dockerFileCommandPrefix := "/bin/sh -c #(nop) "
	createdBy := layer.CreatedBy
	if strings.Contains(layer.CreatedBy, dockerFileCommandPrefix) {
		createdBy = strings.Trim(strings.TrimPrefix(layer.CreatedBy, dockerFileCommandPrefix), " ")
		split := strings.Split(createdBy, " ")
		createdBy = utils.ColoredString(split[0], color.FgYellow) + " " + strings.Join(split[1:], " ")
	}

	createdBy = strings.ReplaceAll(createdBy, "\t", " ")

	size := utils.FormatBinaryBytes(int(layer.Size))
	sizeColor := color.FgWhite
	if size == "0B" {
		sizeColor = color.FgBlue
	}

	return []string{
		utils.ColoredString(id, idColor),
		utils.ColoredString(tag, color.FgGreen),
		utils.ColoredString(size, sizeColor),
		createdBy,
	}
}

// RenderHistory renders the image build history as a table.
func (i *Image) RenderHistory() (string, error) {
	history, err := i.Runtime.ImageHistory(context.Background(), i.ID)
	if err != nil {
		return "", err
	}

	tableBody := lo.Map(history, func(layer domain.ImageHistoryItem, _ int) []string {
		return getHistoryResponseItemDisplayStrings(layer)
	})

	table := make([][]string, 0, 1+len(tableBody))
	table = append(table, []string{"ID", "TAG", "SIZE", "COMMAND"})
	table = append(table, tableBody...)

	return utils.RenderTable(table)
}

// RefreshImages returns the current list of images.
func (c *DockerCommand) RefreshImages() ([]*Image, error) {
	images, err := c.Runtime.ListImages(context.Background())
	if err != nil {
		return nil, err
	}

	ownImages := make([]*Image, len(images))

	for i, img := range images {
		firstTag := ""
		if len(img.RepoTags) > 0 {
			firstTag = img.RepoTags[0]
		}

		nameParts := strings.Split(firstTag, ":")
		tag := ""
		name := "none"
		if len(nameParts) > 1 {
			tag = nameParts[len(nameParts)-1]
			name = strings.Join(nameParts[:len(nameParts)-1], ":")

			for prefix, replacement := range c.Config.UserConfig.Replacements.ImageNamePrefixes {
				if strings.HasPrefix(name, prefix) {
					name = strings.Replace(name, prefix, replacement, 1)
					break
				}
			}
		}

		ownImages[i] = &Image{
			ID:            img.ID,
			Name:          name,
			Tag:           tag,
			Image:         img,
			OSCommand:     c.OSCommand,
			Log:           c.Log,
			DockerCommand: c,
			Runtime:       c.Runtime,
		}
	}

	return ownImages, nil
}

// PruneImages removes dangling images.
func (c *DockerCommand) PruneImages() error {
	_, err := c.Runtime.PruneImages(context.Background())
	return err
}
