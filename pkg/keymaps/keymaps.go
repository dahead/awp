package keymaps

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
)

type KeyDefinition struct {
	DefaultKey string
	Help       string
}

var KeyDefinitions = map[string]KeyDefinition{
	"ShowHelp":           {"ctrl+b", "show/hide commands"},
	"QuitApp":            {"q", "quit"},
	"ToggleStatus":       {"space", "toggle status"},
	"AddTask":            {"a", "add task"},
	"EditTask":           {"e", "edit task"},
	"DeleteTask":         {"d", "delete task"},
	"ToggleViewMode":     {"ctrl+v", "toggle between today's tasks and all tasks"},
	"ShowDoneTasks":      {"ctrl+d", "show only done tasks"},
	"ShowUndoneTasks":    {"ctrl+u", "show only undone tasks"},
	"SearchTasks":        {"ctrl+f", "search tasks"},
	"PrevDay":            {"ctrl+left", "previous day"},
	"NextDay":            {"ctrl+right", "next day"},
	"PrevDayWithTasks":   {"ctrl+shift+left", "previous day with tasks"},
	"NextDayWithTasks":   {"ctrl+shift+right", "next day with tasks"},
	"JumpToToday":        {"h", "jump to today"},
	"ToggleCalendarView": {"ctrl+c", "toggle calendar view"},
	"CalendarLeft":       {"left", "move left in calendar"},
	"CalendarRight":      {"right", "move right in calendar"},
	"CalendarUp":         {"up", "move up in calendar"},
	"CalendarDown":       {"down", "move down in calendar"},
	"CalendarSelect":     {"enter", "select day in calendar"},
	"ToggleSortBy":       {"s", "cycle sort by"},
	"ToggleGroupBy":      {"g", "cycle group by"},
	"ToggleSortOrder":    {"o", "toggle sort order"},
}

type KeyMap struct {
	ShowHelp           key.Binding
	QuitApp            key.Binding
	ToggleStatus       key.Binding
	AddTask            key.Binding
	EditTask           key.Binding
	DeleteTask         key.Binding
	ToggleViewMode     key.Binding
	ShowDoneTasks      key.Binding
	ShowUndoneTasks    key.Binding
	SearchTasks        key.Binding
	PrevDay            key.Binding
	NextDay            key.Binding
	PrevDayWithTasks   key.Binding
	NextDayWithTasks   key.Binding
	JumpToToday        key.Binding
	ToggleCalendarView key.Binding
	CalendarLeft       key.Binding
	CalendarRight      key.Binding
	CalendarUp         key.Binding
	CalendarDown       key.Binding
	CalendarSelect     key.Binding
	ToggleSortBy       key.Binding
	ToggleGroupBy      key.Binding
	ToggleSortOrder    key.Binding
}

func BuildKeyMap(configOverrides map[string]string) KeyMap {
	km := KeyMap{}
	for action, def := range KeyDefinitions {
		keyStr := def.DefaultKey
		if override, exists := configOverrides[action]; exists && override != "" {
			keyStr = override
		}

		switch action {
		case "ShowHelp":
			km.ShowHelp = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "QuitApp":
			km.QuitApp = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "ToggleStatus":
			km.ToggleStatus = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "AddTask":
			km.AddTask = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "EditTask":
			km.EditTask = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "DeleteTask":
			km.DeleteTask = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "ToggleViewMode":
			km.ToggleViewMode = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "ShowDoneTasks":
			km.ShowDoneTasks = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "ShowUndoneTasks":
			km.ShowUndoneTasks = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "SearchTasks":
			km.SearchTasks = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "PrevDay":
			km.PrevDay = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "NextDay":
			km.NextDay = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "PrevDayWithTasks":
			km.PrevDayWithTasks = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "NextDayWithTasks":
			km.NextDayWithTasks = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "JumpToToday":
			km.JumpToToday = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "ToggleCalendarView":
			km.ToggleCalendarView = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "CalendarLeft":
			km.CalendarLeft = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "CalendarRight":
			km.CalendarRight = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "CalendarUp":
			km.CalendarUp = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "CalendarDown":
			km.CalendarDown = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "CalendarSelect":
			km.CalendarSelect = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "ToggleSortBy":
			km.ToggleSortBy = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "ToggleGroupBy":
			km.ToggleGroupBy = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		case "ToggleSortOrder":
			km.ToggleSortOrder = parseKeyBinding(keyStr, def.DefaultKey, def.Help)
		}
	}
	return km
}

func parseKeyBinding(keyStr, defaultKey, helpText string) key.Binding {
	if keyStr == "" {
		keyStr = defaultKey
	}

	// Handle multiple keys separated by commas
	keys := strings.Split(keyStr, ",")
	for i, k := range keys {
		keys[i] = strings.TrimSpace(k)
	}

	return key.NewBinding(
		key.WithKeys(keys...),
		key.WithHelp(keys[0], helpText),
	)
}

// GetDefaultKeyMappings returns the default key mappings for configuration
func GetDefaultKeyMappings() map[string]string {
	keyMappings := make(map[string]string)
	for action, def := range KeyDefinitions {
		keyMappings[action] = def.DefaultKey
	}
	return keyMappings
}
