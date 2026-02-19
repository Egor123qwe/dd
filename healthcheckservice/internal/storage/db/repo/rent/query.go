package rent

const (
	SessionQuery = `SELECT
			s.id, s.total_ram, s.available_ram, s.used_ram, s.price_ram, s.load_speed, s.upload_speed, s.ping, s.price_internet, s.total_price,
			s.created_at as created_at,
			g.id, g.name, g.available_vram, g.used_vram, g.total_vram, g.dlperf, g.price,
			gd.id, gd.name, gd.total_vram,
			cpu.id, cpu.name, cpu.total, cpu.available, cpu.price,
			st.id, st.name, st.type, st.total, st.available, st.used, st.bandwidth, st.price,
            pr.id, pr.template_id
		FROM session as s
		LEFT JOIN gpu as g on s.id = g.session_id
		LEFT JOIN gpu_dict as gd on gd.id = g.gpu_dict_id
		LEFT JOIN cpu on cpu.session_id = s.id
		LEFT JOIN storage as st on st.session_id = s.id
		LEFT JOIN prepull as pr on pr.session_id = s.id
		LEFT JOIN rent as main_rent on main_rent.session_id = s.id
		WHERE s.id = $1 AND s.user_id = $2
`)