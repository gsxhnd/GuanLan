import i18n from "i18next"
import { initReactI18next } from "react-i18next"

import en from "./locales/en.json"
import zh from "./locales/zh.json"

const UI_STORAGE_KEY = "guanlan-ui"

const savedLocale = localStorage.getItem(UI_STORAGE_KEY)
  ? (
      JSON.parse(localStorage.getItem(UI_STORAGE_KEY) ?? "{}") as {
        state?: { locale?: string }
      }
    ).state?.locale
  : undefined

void i18n.use(initReactI18next).init({
  resources: {
    en: { translation: en },
    zh: { translation: zh },
  },
  lng: savedLocale ?? "zh",
  fallbackLng: "en",
  interpolation: {
    escapeValue: false,
  },
})

export default i18n
