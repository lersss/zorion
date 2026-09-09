package models

type StarSystem struct {
    ID     string  `json:"id"`
    Name   string  `json:"name"`
    CoordX float64 `json:"coord_x"`
    CoordY float64 `json:"coord_y"`
}