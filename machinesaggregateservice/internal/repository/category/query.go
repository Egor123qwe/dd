package category

const (
	GPUDictQuery = `
		WITH TotalPrices as (
				SELECT gpu_dict_id, MIN(s.total_price) AS price_min, COUNT(distinct s.id) as total_sessions
				FROM session AS s
				LEFT JOIN gpu AS g ON g.session_id = s.id
				LEFT JOIN gpu_dict AS gd ON gd.id = g.gpu_dict_id
				WHERE s.deleted_at IS NULL AND s.status = $1
				GROUP BY gpu_dict_id
			)
		SELECT distinct gd.id, gd.name, gd.total_vram, tp.total_sessions AS total_sessions, tp.price_min
		FROM gpu_dict AS gd
		LEFT JOIN gpu AS g ON gd.id = g.gpu_dict_id
		INNER JOIN TotalPrices AS tp ON gd.id = tp.gpu_dict_id
		WHERE gd.total_vram BETWEEN $2 AND $3
	`
)
