package models

import "time"

type PlanetResource struct {
    ID              string    `json:"id"`
    PlanetID        string    `json:"planet_id"`
    Name            string    `json:"name"`
    Category        string    `json:"category"`
    Hardness        float64   `json:"hardness"`
    Elasticity      float64   `json:"elasticity"`
    Conductivity    float64   `json:"conductivity"`
    HeatResistance  float64   `json:"heat_resistance"`
    ChemicalActivity float64  `json:"chemical_activity"`
    Density         float64   `json:"density"`
    Biocompatibility float64  `json:"biocompatibility"`
    EnergyDensity   float64   `json:"energy_density"`
    Volatility      float64   `json:"volatility"`
    Quantity        int       `json:"quantity"`
    IsKnown         bool      `json:"is_known"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}