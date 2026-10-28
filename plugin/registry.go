package plugin

import (
	"fmt"
	"reflect"
)

type Entry struct {
	plugin   Plugin
	children []*Entry
}

type registryEntry struct {
	plugin     Plugin
	parentType reflect.Type
}

var registry = make([]registryEntry, 0)

func register(p registryEntry) error { // TODO: cycle detection?
	pluginType := reflect.TypeOf(p.plugin)
	for i := range registry {
		if reflect.TypeOf(registry[i].plugin) == pluginType {
			return fmt.Errorf(
				"plugin %q (%s) is already registered",
				p.plugin.Name(), pluginType,
			)
		}
	}
	registry = append(registry, p)
	return nil
}

func Register(p Plugin) error {
	return register(registryEntry{p, nil})
}

func RegisterWithParent[P Plugin](p Plugin) error {
	parentType := reflect.TypeFor[P]()
	if parentType == reflect.TypeOf(p) {
		return fmt.Errorf("plugin %q cannot be registered as its own parent", p.Name())
	}
	return register(registryEntry{p, parentType})
}

func RegisteredPlugins() ([]*Entry, error) {
	entries := make(map[*Plugin]*Entry)
	root := Entry{nil, make([]*Entry, 0)}

	var visit func(*registryEntry) (*Entry, error)
	visit = func(details *registryEntry) (*Entry, error) {
		if entry, exists := entries[&details.plugin]; exists {
			return entry, nil
		}

		entry := &Entry{details.plugin, make([]*Entry, 0)}
		entries[&details.plugin] = entry

		var parent *Entry
		if details.parentType == nil {
			parent = &root
		} else {
			for i := range registry {
				if reflect.TypeOf(registry[i].plugin) == details.parentType {
					p, err := visit(&registry[i])
					if err != nil {
						return nil, err
					}
					parent = p
					break
				}
			}

			if parent == nil {
				return nil, fmt.Errorf(
					"plugin %q declares parent %s, but no plugin of that type is registered",
					details.plugin.Name(), details.parentType,
				)
			}
		}

		parent.children = append(parent.children, entry)

		return entry, nil
	}

	for i := range registry {
		_, err := visit(&registry[i])
		if err != nil {
			return nil, err
		}
	}

	return root.children, nil
}
