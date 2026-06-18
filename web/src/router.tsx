import { BrowserRouter, Navigate, Route, Routes } from "react-router"

import { AppLayout } from "@/layout"
import { HomePage } from "@/pages/home-page"
import { SettingsPage } from "@/pages/settings-page"

export function AppRouter() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<AppLayout />}>
          <Route index element={<HomePage />} />
          <Route path="settings" element={<SettingsPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
