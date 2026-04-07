package bundle

import "fmt"

// --- Builder ---

// BundleBuilder — покрокова збірка набору корму.
// Дозволяє створювати набір через ланцюжок викликів.
type BundleBuilder struct {
	bundle *Bundle
}

func NewBundleBuilder() *BundleBuilder {
	return &BundleBuilder{
		bundle: &Bundle{
			Extras: []string{},
		},
	}
}

func (b *BundleBuilder) SetName(name string) *BundleBuilder {
	b.bundle.Name = name
	return b
}

func (b *BundleBuilder) SetDogSize(size string) *BundleBuilder {
	b.bundle.DogSize = size
	return b
}

func (b *BundleBuilder) SetFoodType(foodType string) *BundleBuilder {
	b.bundle.FoodType = foodType
	return b
}

func (b *BundleBuilder) AddExtra(extra string) *BundleBuilder {
	b.bundle.Extras = append(b.bundle.Extras, extra)
	return b
}

func (b *BundleBuilder) SetPackSize(size string) *BundleBuilder {
	b.bundle.PackSize = size
	return b
}

// Build — фіналізує набір, повертає помилку якщо обов'язкові поля не заповнені.
func (b *BundleBuilder) Build() (*Bundle, error) {
	if b.bundle.DogSize == "" {
		return nil, fmt.Errorf("dog_size is required")
	}
	if b.bundle.FoodType == "" {
		return nil, fmt.Errorf("food_type is required")
	}
	if b.bundle.PackSize == "" {
		b.bundle.PackSize = "standard"
	}
	if b.bundle.Name == "" {
		b.bundle.Name = fmt.Sprintf("Custom bundle (%s, %s)", b.bundle.DogSize, b.bundle.FoodType)
	}

	result := b.bundle
	b.bundle = &Bundle{Extras: []string{}}
	return result, nil
}
