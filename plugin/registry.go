package plugin

var plugins = make([]Plugin, 0)

func Register(p Plugin) {
	plugins = append(plugins, p)
}

func GetPlugins() []Plugin {
	return plugins
}
