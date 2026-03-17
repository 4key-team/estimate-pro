"use client";

import { useTranslations } from "next-intl";
import { Bell } from "lucide-react";

export default function NotificationsPage() {
  const t = useTranslations("notifications");

  return (
    <div>
      <h1 className="text-2xl font-bold tracking-tight mb-8">{t("title")}</h1>

      <div className="flex flex-col items-center justify-center py-16 text-center">
        <div className="relative mb-6">
          <div className="absolute inset-0 rounded-full bg-gradient-to-br from-amber-500/20 to-orange-500/20 blur-xl" />
          <div className="relative h-16 w-16 rounded-full bg-muted/50 flex items-center justify-center">
            <Bell className="h-7 w-7 text-muted-foreground" />
          </div>
        </div>
        <p className="text-lg font-medium mb-1">{t("empty")}</p>
        <p className="text-sm text-muted-foreground max-w-sm">
          {t("emptyDesc")}
        </p>
      </div>
    </div>
  );
}
