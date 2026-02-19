package tariff

const (
	TariffListQuery = `
		SELECT c.id, c.name as type, COALESCE(COUNT(DISTINCT s.id), 0) AS total_sessions, c.price
		FROM category AS c
		LEFT JOIN gpu_dict as gd on c.id = gd.category_id
		LEFT JOIN gpu as g on gd.id = g.gpu_dict_id
		LEFT JOIN session as s on g.session_id = s.id AND s.deleted_at IS NULL
		WHERE (s.id IS NULL OR s.status = $1) AND (c.name = $2 OR c.name = $3)
		GROUP BY c.id
	`

	SessionsForCategory = `
		SELECT distinct c.id, s.id
		FROM category as c
		LEFT JOIN gpu_dict as gd on c.id = gd.category_id
		LEFT JOIN gpu as g on gd.id = g.gpu_dict_id
		LEFT JOIN session as s on g.session_id = s.id
		WHERE s.status = $1
	`

	GPUDictForTariff = `
		SELECT gd.name, gd.total_vram, gd.category_id
		FROM gpu_dict as gd
		WHERE gd.category_id IS NOT NULL
	`
)
