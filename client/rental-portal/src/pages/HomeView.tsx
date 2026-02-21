import { Link } from 'react-router-dom'
import styles from './HomeView.module.css'

const HERO_IMAGE_URL = 'https://nit-services.com/storage/images/61/92/55/one-78102512.jpg'

export default function HomeView() {
  return (
    <div className={styles.wrap}>
      <div className={styles.hero}>
        <img
          src={HERO_IMAGE_URL}
alt="Аренда вычислительных мощностей"
        className={styles.heroImage}
        />
        <div className={styles.heroOverlay} aria-hidden />
      </div>
      <h1 className={styles.title}>Аренда вычислительных мощностей</h1>
      <p className={styles.subtitle}>Арендуйте мощности у поставщиков или отдайте свои в аренду</p>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>В чём суть</h2>
        <p>
          Сервис связывает тех, у кого есть свободные вычислительные ресурсы (поставщиков), и тех, кому нужны мощности для работы или учёбы (покупателей). Вы можете арендовать вычислительные мощности на время или отдать свои в аренду.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Как пользоваться</h2>
        <p>
          Если вы хотите <strong>арендовать</strong> вычислительные мощности — перейдите в раздел «Покупатель»: выберите поставщика и шаблон окружения, запустите сеанс и подключайтесь по выданным данным.
        </p>
        <p>
          Если вы хотите <strong>предоставлять</strong> свои мощности в аренду — откройте раздел «Поставщик» и настройте доступ к своей машине.
        </p>
      </section>

      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Почему это полезно</h2>
        <p>
          Покупателям не нужно покупать дорогое железо: можно взять нужную конфигурацию в аренду на время, выбрать готовую программу и сразу приступить к работе. Поставщики монетизируют простаивающие мощности и контролируют, когда машина доступна и какие сеансы на ней идут.
        </p>
      </section>

      <div className={styles.links}>
        <Link to="/client" className={styles.linkCard}>
          <span className={styles.linkCardTitle}>Покупатель</span>
          <span className={styles.linkCardDesc}>Оформить аренду, выбрать поставщика и шаблон</span>
        </Link>
        <Link to="/merchant" className={styles.linkCard}>
          <span className={styles.linkCardTitle}>Поставщик</span>
          <span className={styles.linkCardDesc}>Предоставлять свой компьютер в аренду</span>
        </Link>
      </div>
    </div>
  )
}
