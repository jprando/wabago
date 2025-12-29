// config defines the data structures representing the Waybar configuration and its modules.
package config

// WaybarConfig represents the main waybar configuration
type WaybarConfig struct {
	// Bar settings
	Layer          string   `json:"layer,omitempty"`
	Position       string   `json:"position,omitempty"`
	Height         int      `json:"height,omitempty"`
	Width          int      `json:"width,omitempty"`
	Spacing        int      `json:"spacing,omitempty"`
	Margin         string   `json:"margin,omitempty"`
	MarginTop      int      `json:"margin-top,omitempty"`
	MarginBottom   int      `json:"margin-bottom,omitempty"`
	MarginLeft     int      `json:"margin-left,omitempty"`
	MarginRight    int      `json:"margin-right,omitempty"`
	Exclusive      *bool    `json:"exclusive,omitempty"`
	FixedCenter    *bool    `json:"fixed-center,omitempty"`
	Passthrough    *bool    `json:"passthrough,omitempty"`
	GTKLayerShell  *bool    `json:"gtk-layer-shell,omitempty"`
	IPC            *bool    `json:"ipc,omitempty"`
	Mode           string   `json:"mode,omitempty"`
	Output         []string `json:"output,omitempty"`
	Name           string   `json:"name,omitempty"`
	ReloadOnChange *bool    `json:"reload_style_on_change,omitempty"`

	// Module lists
	ModulesLeft   []string `json:"modules-left,omitempty"`
	ModulesCenter []string `json:"modules-center,omitempty"`
	ModulesRight  []string `json:"modules-right,omitempty"`

	// Module configurations (dynamic)
	Modules map[string]interface{} `json:"-"`

	// Raw data for preserving unknown fields
	Raw map[string]interface{} `json:"-"`
}

// GetAllModules returns all modules used in the config
func (c *WaybarConfig) GetAllModules() []string {
	seen := make(map[string]bool)
	var modules []string

	for _, m := range c.ModulesLeft {
		if !seen[m] {
			seen[m] = true
			modules = append(modules, m)
		}
	}
	for _, m := range c.ModulesCenter {
		if !seen[m] {
			seen[m] = true
			modules = append(modules, m)
		}
	}
	for _, m := range c.ModulesRight {
		if !seen[m] {
			seen[m] = true
			modules = append(modules, m)
		}
	}

	return modules
}

// GetModuleConfig returns the configuration for a specific module
func (c *WaybarConfig) GetModuleConfig(name string) map[string]interface{} {
	if config, ok := c.Modules[name].(map[string]interface{}); ok {
		return config
	}
	return make(map[string]interface{})
}

// SetModuleConfig sets the configuration for a specific module
func (c *WaybarConfig) SetModuleConfig(name string, config map[string]interface{}) {
	c.Modules[name] = config
}
