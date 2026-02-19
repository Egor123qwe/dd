package session

const (
	FeedbackList = `
		SELECT s.id, ROUND(AVG(f.score), 2) AS score
		FROM feedback AS f
		JOIN rent AS r ON r.id = f.rent_id
		JOIN session AS s ON r.session_id = s.id
		WHERE s.id IN (?)
		GROUP BY s.id
	`

	SessionQuery = `
			WITH sessions_count AS (
			SELECT s.user_id AS id, COUNT(r.id) AS sessions_count
			FROM rent AS r
			JOIN session AS s ON r.session_id = s.id
			WHERE r.deleted_at IS NOT NULL
			GROUP BY s.user_id
		),

		stopped_sessions AS (
			SELECT s.user_id AS id, COUNT(r.id) AS stopped_sessions
			FROM rent AS r
			JOIN session AS s ON r.session_id = s.id
			WHERE r.deleted_at IS NOT NULL AND r.stop_reason LIKE '%merchant%'
			GROUP BY s.user_id
		)

		SELECT
			s.id, s.total_ram, s.available_ram, s.used_ram, s.price_ram, s.load_speed, s.upload_speed, s.ping, s.price_internet, s.total_price,
			s.created_at AS created_at, ROUND(EXTRACT(EPOCH FROM (now() - s.created_at)))::int as uptime,
			COALESCE((NULLIF(s_c.sessions_count, 0) - ss_c.stopped_sessions::float) / NULLIF(s_c.sessions_count, 0), 0) AS reliability,
			g.id, g.name, g.available_vram, g.used_vram, g.total_vram, g.dlperf, g.price,
			gd.id, gd.name, gd.total_vram,
			cpu.id, cpu.name, cpu.total, cpu.available, cpu.price,
			st.id, st.name, st.type, st.total, st.available, st.used, st.bandwidth, st.price,
						pr.id, pr.template_id,
			u.id, u.avatar
		FROM session AS s
		LEFT JOIN sessions_count AS s_c ON s.user_id = s_c.id
		LEFT JOIN stopped_sessions AS ss_c ON s.user_id = ss_c.id
		LEFT JOIN gpu AS g ON s.id = g.session_id
		LEFT JOIN gpu_dict AS gd ON gd.id = g.gpu_dict_id
		LEFT JOIN cpu ON cpu.session_id = s.id
		LEFT JOIN storage AS st on st.session_id = s.id
		LEFT JOIN prepull AS pr ON pr.session_id = s.id
		LEFT JOIN rent AS main_rent ON main_rent.session_id = s.id
		LEFT JOIN user_profile AS u ON s.user_id = u.user_id
		LEFT JOIN category AS c ON c.id = gd.category_id
		WHERE s.deleted_at IS NULL AND s.status = 'ready'
	`
)
