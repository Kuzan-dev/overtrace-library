package main

// Estructuras del contrato de §6.2. Solo los campos que el validador semántico
// necesita: el schema JSON cubre la forma, esto cubre el significado.

import "encoding/json"

type Source struct {
	URL   string `json:"url"`
	Quote string `json:"quote,omitempty"`
	Kind  string `json:"kind,omitempty"`
}

type Gear struct {
	Name  string `json:"name"`
	Maker string `json:"maker,omitempty"`
	Year  *int   `json:"year,omitempty"`
}

type Asset struct {
	Kind          string  `json:"kind"`
	DisplayName   string  `json:"display_name"`
	SourceURL     string  `json:"source_url"`
	SHA256        *string `json:"sha256,omitempty"`
	License       *string `json:"license,omitempty"`
	Match         string  `json:"match,omitempty"`
	FallbackClass *string `json:"fallback_class,omitempty"`
}

type Node struct {
	ModuleURI   string             `json:"module_uri"`
	FallbackURI *string            `json:"fallback_uri,omitempty"`
	Label       string             `json:"label,omitempty"`
	Enabled     *bool              `json:"enabled,omitempty"`
	Gear        *Gear              `json:"gear,omitempty"`
	Params      map[string]float64 `json:"params,omitempty"`
	Asset       *Asset             `json:"asset,omitempty"`
	Confidence  string             `json:"confidence"`
	Sources     []Source           `json:"sources,omitempty"`
	Note        *string            `json:"note,omitempty"`
	Prefer      []Preferencia      `json:"prefer,omitempty"`
}

// Preferencia acepta las dos formas de `prefer`: la cadena escueta y el objeto
// con los mandos propios del plugin. Se normalizan a una sola aquí para que
// ninguna regla de §9.3 tenga que preguntarse cuál le ha tocado.
type Preferencia struct {
	URI    string
	Params map[string]float64
}

func (p *Preferencia) UnmarshalJSON(b []byte) error {
	var corta string
	if err := json.Unmarshal(b, &corta); err == nil {
		p.URI, p.Params = corta, nil
		return nil
	}
	var larga struct {
		URI    string             `json:"uri"`
		Params map[string]float64 `json:"params"`
	}
	if err := json.Unmarshal(b, &larga); err != nil {
		return err
	}
	p.URI, p.Params = larga.URI, larga.Params
	return nil
}

type Lane struct {
	Label  string   `json:"label,omitempty"`
	GainDB *float64 `json:"gain_db,omitempty"`
	Pan    *float64 `json:"pan,omitempty"`
	Nodes  []Node   `json:"nodes"`
}

type Stage struct {
	Role  string `json:"role"`
	Label string `json:"label,omitempty"`
	Lanes []Lane `json:"lanes"`
}

type Song struct {
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"album,omitempty"`
	ReleaseYear *int   `json:"release_year,omitempty"`
}

type Performance struct {
	Kind        string  `json:"kind"`
	Year        *int    `json:"year,omitempty"`
	Label       string  `json:"label,omitempty"`
	Tuning      *string `json:"tuning,omitempty"`
	CapoFret    *int    `json:"capo_fret,omitempty"`
	StringGauge *string `json:"string_gauge,omitempty"`
}

type Controls struct {
	Volume *float64 `json:"volume,omitempty"`
	Tone   *float64 `json:"tone,omitempty"`
	Coils  *string  `json:"coils,omitempty"`
	Phase  *string  `json:"phase,omitempty"`
}

type SourceInstrument struct {
	Guitar     *string   `json:"guitar,omitempty"`
	PickupType string    `json:"pickup_type"`
	Position   string    `json:"position,omitempty"`
	Controls   *Controls `json:"controls,omitempty"`
	Confidence string    `json:"confidence,omitempty"`
	Sources    []Source  `json:"sources,omitempty"`
}

type Provenance struct {
	Confidence  string   `json:"confidence"`
	Sources     []Source `json:"sources"`
	GeneratedBy *string  `json:"generated_by,omitempty"`
	GeneratedAt *string  `json:"generated_at,omitempty"`
}

type TunedFor struct {
	GuitarID    string  `json:"guitar_id"`
	DisplayName *string `json:"display_name,omitempty"`
	PositionID  *string `json:"position_id,omitempty"`
}

type Preset struct {
	ID               string            `json:"id"`
	SchemaVersion    int               `json:"schema_version"`
	ParentID         *string           `json:"parent_id,omitempty"`
	Song             Song              `json:"song"`
	Section          string            `json:"section"`
	Performance      Performance       `json:"performance"`
	SourceInstrument *SourceInstrument `json:"source_instrument,omitempty"`
	TunedFor         *TunedFor         `json:"tuned_for,omitempty"`
	Name             string            `json:"name,omitempty"`
	Stages           []Stage           `json:"stages"`
	Provenance       Provenance        `json:"provenance"`
}

// ── catálogo de módulos ──────────────────────────────────────────────────────

type Param struct {
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Default float64 `json:"default"`
	Unit    string  `json:"unit"`
	Taper   string  `json:"taper"`
}

type Module struct {
	URI         string           `json:"uri"`
	Tier        string           `json:"tier"`
	Category    string           `json:"category"`
	Label       string           `json:"label"`
	IsAmp       bool             `json:"is_amp"`
	EngineOwned bool             `json:"engine_owned"`
	AssetKind   string           `json:"asset_kind"`
	Params      map[string]Param `json:"params"`
}

type Catalog struct {
	Version int      `json:"catalog_version"`
	Modules []Module `json:"modules"`
}
