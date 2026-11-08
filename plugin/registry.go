package plugin

import (
	"fmt"
	"reflect"
)

type Entry struct {
	loader   Loader
	children []*Entry
}

type registryEntry struct {
	loader     Loader
	parentType reflect.Type
}

var registry = make([]registryEntry, 0)

func register(p registryEntry) error { // TODO: cycle detection?
	loaderType := reflect.TypeOf(p.loader)
	for i := range registry {
		if reflect.TypeOf(registry[i].loader) == loaderType {
			return fmt.Errorf(
				"plugin loader %q (%s) is already registered",
				p.loader.Name(), loaderType,
			)
		}
	}
	registry = append(registry, p)
	return nil
}

func Register(p Loader) error {
	return register(registryEntry{p, nil})
}

func RegisterWithParent[L Loader](p Loader) error {
	parentType := reflect.TypeFor[L]()
	if parentType == reflect.TypeOf(p) {
		return fmt.Errorf("plugin loader %q cannot be registered as its own parent", p.Name())
	}
	return register(registryEntry{p, parentType})
}

func RegisteredLoaders() ([]*Entry, error) {
	entries := make(map[*Loader]*Entry)
	root := Entry{nil, make([]*Entry, 0)}

	var visit func(*registryEntry) (*Entry, error)
	visit = func(details *registryEntry) (*Entry, error) {
		if entry, exists := entries[&details.loader]; exists {
			return entry, nil
		}

		entry := &Entry{details.loader, make([]*Entry, 0)}
		entries[&details.loader] = entry

		var parent *Entry
		if details.parentType == nil {
			parent = &root
		} else {
			for i := range registry {
				if reflect.TypeOf(registry[i].loader) == details.parentType {
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
					"plugin loader %q declares parent %s, but no plugin of that type is registered",
					details.loader.Name(), details.parentType,
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
