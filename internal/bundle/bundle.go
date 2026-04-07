package bundle

import "fmt"

// Bundle — набір корму, який користувач може зібрати або скопіювати з шаблону.
type Bundle struct {
	Name       string   `json:"name"`
	DogSize    string   `json:"dog_size"`    // small, medium, large
	FoodType   string   `json:"food_type"`   // dry, wet, mixed
	Extras     []string `json:"extras"`      // vitamins, toy, bowl
	PackSize   string   `json:"pack_size"`   // standard, large, family
}

func (b *Bundle) String() string {
	return fmt.Sprintf("Bundle{%s, size=%s, food=%s, extras=%v, pack=%s}",
		b.Name, b.DogSize, b.FoodType, b.Extras, b.PackSize)
}

// --- Prototype ---

// Clone створює глибоку копію набору, яку можна змінювати незалежно від оригіналу.
func (b *Bundle) Clone() *Bundle {
	extrasCopy := make([]string, len(b.Extras))
	copy(extrasCopy, b.Extras)

	return &Bundle{
		Name:     b.Name,
		DogSize:  b.DogSize,
		FoodType: b.FoodType,
		Extras:   extrasCopy,
		PackSize: b.PackSize,
	}
}
