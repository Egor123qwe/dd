package category

type GPUDict struct {
	ID    string  `json:"id,omitempty" db:"id"`
	Name  string  `json:"name,omitempty" db:"name"`
	Price float64 `json:"price,omitempty" db:"price"`
}

type CPUDict struct {
	ID              int64   `json:"id" db:"id"`
	Name            string  `json:"name" db:"name"`
	PricePerMinute  float64 `json:"price_per_minute" db:"price_per_minute"`
}

type PricingConfig struct {
	BasePerMinute                float64 `db:"base_per_minute"`
	RAMPerGBPerMinute            float64 `db:"ram_per_gb_per_minute"`
	StorageHDDPerGBPerMinute     float64 `db:"storage_hdd_per_gb_per_minute"`
	StorageSSDPerGBPerMinute     float64 `db:"storage_ssd_per_gb_per_minute"`
	InternetPerMbitPerMinute     float64 `db:"internet_per_mbit_per_minute"`
}
