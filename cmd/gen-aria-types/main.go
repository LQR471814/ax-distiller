package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
)

type JsonRole struct {
	Name        string   `json:"name"`
	ParentRoles []string `json:"parentRoles"`
}

type Role struct {
	Name     string
	IsRoot   bool
	Children []*Role
}

func printRoleTree(role *Role, depth int) string {
	result := ""
	for i := 0; i < depth; i++ {
		result += "  "
	}
	result += "* " + role.Name + "\n"
	for _, c := range role.Children {
		result += printRoleTree(c, depth+1)
	}
	return result
}

type categories = map[string][]string

func generateCategories(role *Role, out categories, targets []string, parent string) {
	currentParent := parent
	if slices.Contains(targets, role.Name) {
		currentParent = role.Name
	}

	if !slices.Contains(out[role.Name], currentParent) {
		out[role.Name] = append([]string{currentParent}, out[role.Name]...)
	}

	for _, c := range role.Children {
		generateCategories(c, out, targets, currentParent)
	}
}

func generateCategorizationCode(targets []string, rolecats categories) string {
	categoryType := "type Category = int"

	categoryIdMap := map[string]string{}
	targetEnum := "const (\n"
	for i, t := range targets {
		id := "CATEGORY_" + strings.ToUpper(t)
		categoryIdMap[t] = id

		if i == 0 {
			targetEnum += fmt.Sprintf("  %s Category = iota\n", id)
			continue
		}
		targetEnum += fmt.Sprintf("  %s\n", id)
	}
	targetEnum += ")"

	initializer := "\n"
	for role, cats := range rolecats {
		catList := ""
		for i, c := range cats {
			catList += categoryIdMap[c]
			if i < len(cats)-1 {
				catList += ", "
			}
		}
		initializer += fmt.Sprintf("    \"%s\": {%s},\n", role, catList)
	}
	mapping := fmt.Sprintf("var RoleCategoryMap = map[string][]Category{%s}", string(initializer))

	return fmt.Sprintf("%s\n\n%s\n\n%s", categoryType, targetEnum, mapping)
}

func main() {
	buff, err := os.ReadFile("./role_info.json")
	if err != nil {
		log.Fatal(err)
	}

	referenceRoles := map[string]JsonRole{}
	err = json.Unmarshal(buff, &referenceRoles)
	if err != nil {
		log.Fatal(err)
	}

	roles := map[string]*Role{}
	for _, ref := range referenceRoles {
		current, ok := roles[ref.Name]
		if !ok {
			current = &Role{Name: ref.Name}
			roles[ref.Name] = current
		}
		if len(ref.ParentRoles) == 0 {
			current.IsRoot = true
		}

		for _, parentName := range ref.ParentRoles {
			parentRole, ok := roles[parentName]
			if !ok {
				parentRole = &Role{
					Name: parentName,
				}
				roles[parentName] = parentRole
			}
			parentRole.Children = append(parentRole.Children, current)
		}
	}

	// for _, r := range roles {
	// 	if r.IsRoot {
	// 		fmt.Print(printRoleTree(r, 0))
	// 	}
	// }

	targets := []string{
		"dialog",
		"widget",
		"document",
		"landmark",
		"structure",
		"section",
		"sectionhead",
		"generic",
	}
	cats := categories{}
	generateCategories(roles["roletype"], cats, targets, "roletype")
	fmt.Println(generateCategorizationCode(targets, cats))
}
