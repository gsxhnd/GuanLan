import { useTranslation } from "react-i18next"

import {
  ContentCard,
  LogButton,
  PageHeader,
  StatCard,
  StatGrid,
  StatusBadge,
} from "@/components/page"
import { PreferencesPanel } from "@/components/settings/preferences-panel"
import { Button } from "@/components/ui/button"
import { useToast } from "@/hooks/use-toast"

export function SystemPage() {
  const { t } = useTranslation()
  const { showToast, Toast } = useToast()

  return (
    <div className="flex flex-col gap-6">
      {Toast}
      <PageHeader
        title={t("pages.system.title")}
        description={t("pages.system.description")}
      />

      <StatGrid cols={3}>
        <StatCard label={t("settings.services.api")}>
          <div className="mt-2">
            <StatusBadge variant="success" dot>
              healthy
            </StatusBadge>
          </div>
          <p className="mt-2 text-xs text-muted-foreground">
            {t("settings.services.apiUptime")}
          </p>
        </StatCard>
        <StatCard label={t("settings.services.duckdb")}>
          <div className="mt-2">
            <StatusBadge variant="success" dot>
              healthy
            </StatusBadge>
          </div>
          <p className="mt-2 font-mono text-xs text-muted-foreground">
            ~/guanlan/data.duckdb
          </p>
        </StatCard>
        <StatCard label={t("settings.services.python")}>
          <div className="mt-2">
            <StatusBadge variant="warn" dot>
              degraded
            </StatusBadge>
          </div>
          <p className="mt-2 text-xs text-muted-foreground">
            {t("settings.services.pythonDetail")}
          </p>
        </StatCard>
      </StatGrid>

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <PreferencesPanel />

        <ContentCard title={t("settings.deploymentTitle")}>
          <dl className="grid grid-cols-[auto_1fr] gap-x-6 gap-y-3 text-sm">
            <dt className="text-muted-foreground">
              {t("settings.deployment.dbPath")}
            </dt>
            <dd className="m-0 font-mono text-[13px]">~/guanlan/data.duckdb</dd>
            <dt className="text-muted-foreground">
              {t("settings.deployment.logDir")}
            </dt>
            <dd className="m-0 font-mono text-[13px]">~/guanlan/logs/</dd>
            <dt className="text-muted-foreground">
              {t("settings.deployment.scheduler")}
            </dt>
            <dd className="m-0">{t("settings.deployment.schedulerValue")}</dd>
            <dt className="text-muted-foreground">
              {t("settings.deployment.mode")}
            </dt>
            <dd className="m-0">{t("settings.deployment.modeValue")}</dd>
          </dl>
        </ContentCard>
      </div>

      <ContentCard title={t("settings.shortcutsTitle")}>
        <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap">
          <LogButton
            size="default"
            onClick={() => showToast("日志入口 · Phase 2 接入")}
          />
          <Button variant="secondary">
            {t("settings.shortcuts.dataVersions")}
          </Button>
          <Button variant="ghost">{t("settings.shortcuts.questions")}</Button>
        </div>
      </ContentCard>
    </div>
  )
}
