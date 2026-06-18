import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"

export function HomePage() {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col gap-8">
      <section className="flex flex-col gap-3">
        <h1 className="text-2xl font-semibold">{t("home.title")}</h1>
        <p className="max-w-2xl text-sm text-muted-foreground">
          {t("home.description")}
        </p>
        <div className="flex flex-wrap gap-2 pt-2">
          <Button>{t("home.primaryAction")}</Button>
          <Button variant="outline">{t("home.secondaryAction")}</Button>
        </div>
      </section>

      <section className="grid gap-3 md:grid-cols-3">
        {["first", "second", "third"].map((item) => (
          <article
            key={item}
            className="flex flex-col gap-2 rounded-lg border p-4"
          >
            <p className="text-sm font-medium">
              {t(`home.cards.${item}.title`)}
            </p>
            <p className="text-sm text-muted-foreground">
              {t(`home.cards.${item}.description`)}
            </p>
          </article>
        ))}
      </section>
    </div>
  )
}
