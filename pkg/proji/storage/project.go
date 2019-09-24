package storage

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/nikoksr/proji/pkg/helper"

	"github.com/otiai10/copy"
)

// Project struct represents a proji project
type Project struct {
	// The project name
	Name string

	// The template class
	Class *Class

	// The install path for the project
	InstallPath string

	// The class label
	label string

	// The storage service
	store Service

	// The original working directory
	owd string
}

// NewProject returns a new project
func NewProject(name, label, cwd string, store Service) (*Project, error) {
	// Validate label
	id, err := store.DoesLabelExist(label)
	if err != nil {
		return nil, err
	}

	// Load class by id
	class, err := store.LoadClassByID(id)
	if err != nil {
		return nil, err
	}

	// Append a slash if not exists. Out of convenience.
	if cwd[:len(cwd)-1] != "/" {
		cwd += "/"
	}

	return &Project{
		Name:        name,
		Class:       class,
		InstallPath: cwd + name,
		label:       label,
		owd:         cwd,
	}, nil
}

// Create starts the creation of a project.
func (proj *Project) Create() error {
	// Create the project folder
	fmt.Println("> Creating project folder...")
	if err := proj.createProjectFolder(); err != nil {
		return err
	}

	// Chdir into the new project folder and defer chdir back to old cwd
	if err := os.Chdir(proj.Name); err != nil {
		return err
	}
	defer os.Chdir(proj.owd)

	// Create subfolders
	fmt.Println("> Creating subfolders...")
	if err := proj.createSubFolders(); err != nil {
		return err
	}

	// Create files
	fmt.Println("> Creating files...")
	if err := proj.createFiles(); err != nil {
		return err
	}

	// Copy templates
	fmt.Println("> Copying templates...")
	if err := proj.copyTemplates(); err != nil {
		return err
	}

	// Run scripts
	fmt.Println("> Running scripts...")
	return proj.runScripts()
}

// createProjectFolder tries to create the main project folder.
func (proj *Project) createProjectFolder() error {
	return os.Mkdir(proj.Name, os.ModePerm)
}

func (proj *Project) createSubFolders() error {
	// Regex to replace keyword with project name
	re := regexp.MustCompile(`__PROJECT_NAME__`)

	for folder, template := range proj.Class.Folders {
		// Skip, folder has a template
		if len(template) > 0 {
			continue
		}

		// Replace keyword with project name
		folder = re.ReplaceAllString(folder, proj.Name)
		if err := os.MkdirAll(folder, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func (proj *Project) createFiles() error {
	re := regexp.MustCompile(`__PROJECT_NAME__`)

	for file, template := range proj.Class.Files {
		// Skip, file has a template
		if len(template) > 0 {
			continue
		}

		// Replace keyword with project name
		file = re.ReplaceAllString(file, proj.Name)
		if _, err := os.OpenFile(file, os.O_RDONLY|os.O_CREATE, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func (proj *Project) copyTemplates() error {
	re := regexp.MustCompile(`__PROJECT_NAME__`)

	for _, templates := range []map[string]string{proj.Class.Folders, proj.Class.Files} {
		for fifo, template := range templates {
			if len(template) < 1 {
				continue
			}

			// Replace keyword with project name
			fifo = re.ReplaceAllString(fifo, proj.Name)
			template = helper.GetConfigDir() + "/templates/" + template
			if err := copy.Copy(template, fifo); err != nil {
				return err
			}
		}
	}
	return nil
}

func (proj *Project) runScripts() error {
	for script, runAsSudo := range proj.Class.Scripts {
		script = helper.GetConfigDir() + "/scripts/" + script

		if runAsSudo {
			script = "sudo " + script
		}

		cmd := exec.Command(script)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}