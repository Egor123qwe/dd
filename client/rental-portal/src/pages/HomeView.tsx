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
        <h2 className={styles.sectionTitle}>Что можно делать</h2>
        <p>
          Просматривать характеристики доступных машин (память, процессор, сеть, диск), выбирать готовые шаблоны с операционной системой и программами, запускать сеансы аренды и подключаться к удалённому доступу. Поставщики могут управлять своими ресурсами и сеансами.
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
