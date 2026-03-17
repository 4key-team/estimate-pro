"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Bell, Globe, User } from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/hero-195-1";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ThemeToggle } from "@/components/ui/theme-toggle";

export default function SettingsPage() {
  const t = useTranslations();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");

  useEffect(() => {
    setName(localStorage.getItem("user_name") ?? "");
    setEmail(localStorage.getItem("user_email") ?? "");
  }, []);

  const handleSave = () => {
    localStorage.setItem("user_name", name);
    localStorage.setItem("user_email", email);
  };

  return (
    <div>
      <h1 className="text-3xl font-bold tracking-tight mb-8 max-w-2xl mx-auto">
        {t("dashboard.settings")}
      </h1>

      <div className="space-y-6 max-w-2xl mx-auto">
        {/* Profile */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <User className="h-5 w-5 text-muted-foreground" />
              <CardTitle className="text-lg">{t("settings.profile")}</CardTitle>
            </div>
            <CardDescription>
              {t("settings.profileDesc")}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label>{t("auth.name")}</Label>
              <Input
                placeholder={t("settings.namePlaceholder")}
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label>{t("auth.email")}</Label>
              <Input
                type="email"
                placeholder="email@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
            <Button onClick={handleSave}>{t("common.save")}</Button>
          </CardContent>
        </Card>

        {/* Appearance */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Globe className="h-5 w-5 text-muted-foreground" />
              <CardTitle className="text-lg">{t("settings.appearance")}</CardTitle>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">{t("settings.theme")}</p>
                <p className="text-sm text-muted-foreground">
                  {t("settings.themeDesc")}
                </p>
              </div>
              <ThemeToggle />
            </div>
          </CardContent>
        </Card>

        {/* Notifications */}
        <Card>
          <CardHeader>
            <div className="flex items-center gap-2">
              <Bell className="h-5 w-5 text-muted-foreground" />
              <CardTitle className="text-lg">
                {t("dashboard.notifications")}
              </CardTitle>
            </div>
            <CardDescription>
              {t("settings.notificationsDesc")}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-sm text-muted-foreground py-4 text-center">
              {t("settings.comingSoon")}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
