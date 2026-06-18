import { useTranslation } from "react-i18next"

type SectionPageProps = {
  titleKey: string
  descriptionKey: string
}

export function SectionPage({ titleKey, descriptionKey }: SectionPageProps) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col gap-3">
      <h1 className="text-2xl font-semibold">{t(titleKey)}</h1>
      <p className="max-w-2xl text-sm text-muted-foreground">
        {t(descriptionKey)}
      </p>
    </div>
  )
}
