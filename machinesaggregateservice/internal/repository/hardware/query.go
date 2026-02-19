package hardware

const (
	GPUListQuery = `
		SELECT
			g.id, g.name, g.available_vram, g.used_vram, g.total_vram ,g.dlperf, g.price
		FROM gpu AS g
			INNER JOIN session as s on g.session_id = s.id
			INNER JOIN rent as r on r.session_id = s.id and r.status NOT IN ('started', 'pending')
			WHERE s.deleted_at IS NULL AND (r.session_id IS NULL OR r.status NOT IN ('started', 'pending'))
	`
)
