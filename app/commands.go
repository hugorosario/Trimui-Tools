package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hugorosario/trimuitools/shell"
)

func (h *MainScreen) RunCommandCaptured(item *menuItem) {
	if item.Execute == "" {
		return
	}
	h.showLineView = item.Output == "line"
	h.showShellView = item.Output == "full"
	h.showDialogView = false
	if h.showLineView || h.showShellView {
		h.title = item.DisplayLabel
	}

	cmd := item.Execute
	if (item.Type == "select") || (item.Type == "toggle") {
		cmd = fmt.Sprintf("%s %d", cmd, item.Selected)
	}
	h.shellView.Clear()
	h.lineView.SetText("")
	h.dialogView.SetText("")
	h.shellInput, h.shellCancel, _ = shell.RunCommandAsync(cmd, func(event shell.ShellEvent) {
		switch event.Type {
		case "start":
			h.shellRunnning = true
			h.shellView.AddText("Executing: " + event.Data)
		case "output":
			// if there are tags in between {{ and }}, ignore this output since it is only applicable for the "load" commands
			if strings.Contains(event.Data, "{{") && strings.Contains(event.Data, "}}") {
				return
			}
			data := strings.ReplaceAll(event.Data, "\\n", "\n")
			//example of confirm dialog: confirm:Are you sure you want to continue?
			//check if prefix is for showing a confirm dialog
			if strings.HasPrefix(data, "confirm:") {
				h.dialogView.SetText(strings.TrimPrefix(data, "confirm:"))
				h.title = item.DisplayLabel
				h.showDialogView = true
			} else if h.showLineView {
				h.lineView.SetText(data)
			}
			h.shellView.AddText(data)
		case "error":
			h.shellView.AddText(event.Data)
		case "exit":
			h.showDialogView = false
			h.shellView.AddText("Exit code: " + event.Data)
			exitCode, err := strconv.Atoi(event.Data)
			if err != nil {
				h.shellView.AddText("Error parsing exit code: " + err.Error())
				exitCode = 1
			}
			if exitCode != 0 {
				h.showLineView = false
				h.showShellView = true
			}
			if h.shellCancel != nil {
				h.shellCancel()
			}
			if (item.Type == "toggle") && (exitCode != 0) {
				item.Selected = 1 - item.Selected
			}

			h.runLoadCommands(item)
		case "end":
			h.shellRunnning = false
			h.title = h.lists[len(h.lists)-1].Title
			//run load commands for all menu items
			for _, item := range h.GetList().GetItems() {
				h.runLoadCommands(item.Value)
			}
		}
	})
}

func (h *MainScreen) runLoadCommands(item *menuItem) {
	item.DisplayLabel = item.Label
	item.DisplayDescription = item.Description
	labelTags := getTags(item.Label)
	descriptionTags := getTags(item.Description)

	if item.Load != "" {
		cmd := item.Load
		output, err := shell.RunCommandSync(cmd)
		if err == nil {
			//iterate all lines in the output
			for _, line := range strings.Split(output, "\n") {
				//find if the line contains any tag values in the format of {{key=value}}
				tags := getTags(line)
				for _, tag := range tags {
					//split the tag into key and value
					parts := strings.Split(tag, "=")
					if len(parts) != 2 {
						continue
					}
					key := parts[0]
					value := parts[1]

					if (strings.HasPrefix(key, "state")) && (strings.Contains("toggle,select", item.Type)) {
						//when state tag is meant for the item itself, the key contains only "state"
						//when its meant for another item the key contains the item path like state(0.1.2),
						//ex: state(0.1)=1 for setting the state=1 of the second item in the first submenu
						state, err := strconv.Atoi(value)
						if err != nil {
							continue
						}
						//example path: state(0.1)=1
						if strings.Contains(key, "(") {
							if (len(h.lists) > 0) && (len(h.lists[0].List.GetValues()) > 0) {
								path := strings.Split(key, "(")[1]
								path = strings.Trim(path, ")")
								//find the other in the path
								other := findItem(h.lists[0].List.GetValues(), path)
								if other != nil {
									other.Selected = state
								}
							}
						} else {
							item.Selected = state
						}
					}

					if strings.HasPrefix(key, "disable") {
						//when tag is meant for the item itself the key contains only "disable"
						//when its meant for another item the key contains the item path like disable(0.1.2),
						//ex: disable(0.1)=1 for setting the disable=1 of the second item in the first submenu
						disabled, err := strconv.Atoi(value)
						if err != nil {
							continue
						}
						//example path: disable(0.1)=1
						if strings.Contains(key, "(") {
							if (len(h.lists) > 0) && (len(h.lists[0].List.GetValues()) > 0) {
								path := strings.Split(key, "(")[1]
								path = strings.Trim(path, ")")
								//find the other in the path
								other := findItem(h.lists[0].List.GetValues(), path)
								if other != nil {
									other.Disabled = disabled == 1
								}
							}
						} else {
							item.Disabled = disabled == 1
						}
					}

					//replace the tag in the label and description
					for i, labelTag := range labelTags {
						if strings.Contains(labelTag, key) {
							item.DisplayLabel = strings.ReplaceAll(item.DisplayLabel, "{{"+labelTag+"}}", value)
							labelTags = append(labelTags[:i], labelTags[i+1:]...)
						}
					}

					for i, descriptionTag := range descriptionTags {
						if strings.Contains(descriptionTag, key) {
							item.DisplayDescription = strings.ReplaceAll(item.DisplayDescription, "{{"+descriptionTag+"}}", value)
							descriptionTags = append(descriptionTags[:i], descriptionTags[i+1:]...)
						}
					}
				}
			}
		}
	}

	//replace any remaining tags with empty strings
	for _, labelTag := range labelTags {
		item.DisplayLabel = strings.ReplaceAll(item.DisplayLabel, "{{"+labelTag+"}}", "")
	}

	for _, descriptionTag := range descriptionTags {
		item.DisplayDescription = strings.ReplaceAll(item.DisplayDescription, "{{"+descriptionTag+"}}", "")
	}

	//recursively run the load commands for all menu items
	for _, subitem := range item.MenuItems {
		h.runLoadCommands(subitem)
	}
}
