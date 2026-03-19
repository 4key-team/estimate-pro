"use client";

import { useMemo } from "react";
import { useTranslations } from "next-intl";
import { useQuery, useQueries } from "@tanstack/react-query";
import { listProjects } from "@/features/projects/api";
import { listDocuments, type Document as DocType } from "@/features/documents/api";
import { listEstimations, type Estimation } from "@/features/estimation/api";
import { ActivityLogsTable, type ActivityLog } from "@/features/activity/components/activity-logs-table";

export default function NotificationsPage() {
  const t = useTranslations();
  const { data: projectsData } = useQuery({
    queryKey: ["projects"],
    queryFn: () => listProjects(),
  });

  const projects = projectsData?.projects ?? [];

  const docQueries = useQueries({
    queries: projects.map((p) => ({
      queryKey: ["documents", p.id],
      queryFn: () => listDocuments(p.id),
    })),
  });

  const estQueries = useQueries({
    queries: projects.map((p) => ({
      queryKey: ["estimations", p.id],
      queryFn: () => listEstimations(p.id),
    })),
  });

  const logs: ActivityLog[] = useMemo(() => {
    const items: ActivityLog[] = [];
    let counter = 0;

    projects.forEach((p, idx) => {
      (docQueries[idx]?.data ?? []).forEach((d: DocType) => {
        items.push({
          id: String(++counter),
          timestamp: d.created_at,
          level: "success",
          service: p.name,
          message: `Загружен документ: ${d.title}`,
          status: "uploaded",
          tags: ["document", p.name],
        });
      });
      (estQueries[idx]?.data ?? []).forEach((e: Estimation) => {
        items.push({
          id: String(++counter),
          timestamp: e.created_at,
          level: e.status === "submitted" ? "info" : "warning",
          service: p.name,
          message: e.status === "submitted" ? "Оценка отправлена" : "Оценка создана (черновик)",
          status: e.status === "submitted" ? "submitted" : "draft",
          tags: ["estimation", p.name],
        });
      });
    });

    return items.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
  }, [projects, docQueries, estQueries]);

  return (
    <div>
      <h1 className="text-2xl font-bold tracking-tight mb-8">{t("notifications.title")}</h1>
      <ActivityLogsTable logs={logs} />
    </div>
  );
}
