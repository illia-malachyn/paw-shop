package bundle

// BundleRegistry — реєстр шаблонних наборів (прототипів).
// Користувач може отримати копію шаблону і змінити під себе.
type BundleRegistry struct {
	templates map[string]*Bundle
}

func NewBundleRegistry() *BundleRegistry {
	r := &BundleRegistry{
		templates: make(map[string]*Bundle),
	}
	r.loadDefaults()
	return r
}

func (r *BundleRegistry) loadDefaults() {
	r.templates["puppy"] = &Bundle{
		Name:     "Для цуценят",
		DogSize:  "small",
		FoodType: "wet",
		Extras:   []string{"vitamins", "toy"},
		PackSize: "standard",
	}
	r.templates["large_breed"] = &Bundle{
		Name:     "Для великих порід",
		DogSize:  "large",
		FoodType: "dry",
		Extras:   []string{"vitamins"},
		PackSize: "family",
	}
	r.templates["senior"] = &Bundle{
		Name:     "Для старших собак",
		DogSize:  "medium",
		FoodType: "mixed",
		Extras:   []string{"vitamins", "bowl"},
		PackSize: "standard",
	}
}

// Get — повертає клон шаблону за ключем.
func (r *BundleRegistry) Get(key string) *Bundle {
	t, ok := r.templates[key]
	if !ok {
		return nil
	}
	return t.Clone()
}

// List — повертає список доступних шаблонів (ключ + назва).
func (r *BundleRegistry) List() map[string]string {
	result := make(map[string]string)
	for key, b := range r.templates {
		result[key] = b.Name
	}
	return result
}
