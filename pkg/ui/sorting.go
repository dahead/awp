package ui

import (
	"fmt"
	"sort"
	"strings"

	"awp/pkg/database"
)

// GroupedTasks represents tasks grouped by a common attribute
type GroupedTasks struct {
	GroupName string
	Tasks     []database.TodoItem
}

// SortTasks sorts tasks based on the specified criteria
func (m *Model) SortTasks(tasks []database.TodoItem) []database.TodoItem {
	sortedTasks := make([]database.TodoItem, len(tasks))
	copy(sortedTasks, tasks)

	sort.Slice(sortedTasks, func(i, j int) bool {
		var result bool

		switch m.sortBy {
		case database.SortByTitle:
			result = strings.ToLower(sortedTasks[i].Title) < strings.ToLower(sortedTasks[j].Title)
		case database.SortByDescription:
			result = strings.ToLower(sortedTasks[i].Description) < strings.ToLower(sortedTasks[j].Description)
		case database.SortByDueDate:
			result = sortedTasks[i].DueDate.Before(sortedTasks[j].DueDate)
		case database.SortByCreated:
			result = sortedTasks[i].Created.Before(sortedTasks[j].Created)
		case database.SortByStatus:
			result = !sortedTasks[i].Status && sortedTasks[j].Status // Undone first
		case database.SortByProject:
			proj1 := getFirstProject(sortedTasks[i])
			proj2 := getFirstProject(sortedTasks[j])
			result = strings.ToLower(proj1) < strings.ToLower(proj2)
		case database.SortByContext:
			ctx1 := getFirstContext(sortedTasks[i])
			ctx2 := getFirstContext(sortedTasks[j])
			result = strings.ToLower(ctx1) < strings.ToLower(ctx2)
		}

		if m.sortOrder == database.SortDesc {
			result = !result
		}
		return result
	})

	return sortedTasks
}

// GroupTasks groups tasks based on the specified criteria
func (m *Model) GroupTasks(tasks []database.TodoItem) []GroupedTasks {
	if m.groupBy == database.GroupByNone {
		return []GroupedTasks{{GroupName: "", Tasks: m.SortTasks(tasks)}}
	}

	groups := make(map[string][]database.TodoItem)

	for _, task := range tasks {
		var groupKey string

		switch m.groupBy {
		case database.GroupByProject:
			groupKey = getFirstProject(task)
			if groupKey == "" {
				groupKey = "No Project"
			} else {
				groupKey = "+" + groupKey
			}

		case database.GroupByContext:
			groupKey = getFirstContext(task)
			if groupKey == "" {
				groupKey = "No Context"
			} else {
				groupKey = "@" + groupKey
			}

		case database.GroupByDueDateDaily:
			groupKey = task.DueDate.Format("2006-01-02")

		case database.GroupByDueDateWeekly:
			year, week := task.DueDate.ISOWeek()
			groupKey = fmt.Sprintf("Week %d, %d", week, year)

		case database.GroupByDueDateMonthly:
			groupKey = task.DueDate.Format("January 2006")

		case database.GroupByDueDateYearly:
			groupKey = task.DueDate.Format("2006")
		}

		groups[groupKey] = append(groups[groupKey], task)
	}

	// Convert map to sorted slice
	var result []GroupedTasks
	var groupNames []string
	for name := range groups {
		groupNames = append(groupNames, name)
	}
	sort.Strings(groupNames)

	for _, name := range groupNames {
		result = append(result, GroupedTasks{
			GroupName: name,
			Tasks:     m.SortTasks(groups[name]),
		})
	}

	return result
}

// Helper functions
func getFirstProject(task database.TodoItem) string {
	if len(task.Projects) > 0 {
		return task.Projects[0]
	}
	return ""
}

func getFirstContext(task database.TodoItem) string {
	if len(task.Contexts) > 0 {
		return task.Contexts[0]
	}
	return ""
}
