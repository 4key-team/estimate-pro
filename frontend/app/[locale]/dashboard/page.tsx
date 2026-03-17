"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { useQuery } from "@tanstack/react-query";
import {
  Plus,
  FolderKanban,
  Users,
  FileText,
  BarChart3,
  Sparkles,
  FolderOpen,
} from "lucide-react";
import { Link } from "@/i18n/navigation";
import { listWorkspaces, listProjects } from "@/features/projects/api";
import { CreateProjectDialog } from "@/features/projects/components/create-project-dialog";

export default function DashboardPage() {
  const t = useTranslations();
  const [name, setName] = useState("");
  const [createOpen, setCreateOpen] = useState(false);

  useEffect(() => {
    setName(localStorage.getItem("user_name") ?? "");
  }, []);

  const greeting = name
    ? t("dashboard.welcomeBack", { name })
    : t("dashboard.welcome");

  // Fetch workspaces
  const { data: workspaces } = useQuery({
    queryKey: ["workspaces"],
    queryFn: listWorkspaces,
    retry: false,
  });

  const workspaceId = workspaces?.[0]?.id;

  // Fetch projects for first workspace
  const { data: projectsData } = useQuery({
    queryKey: ["projects", workspaceId],
    queryFn: () => listProjects(workspaceId!),
    enabled: !!workspaceId,
  });

  const projects = projectsData?.projects ?? [];
  const projectCount = projectsData?.meta?.total ?? 0;

  const stats = [
    {
      icon: FolderKanban,
      value: String(projectCount),
      label: t("projects.title"),
      color: "text-violet-500",
      bg: "bg-violet-500/10",
    },
    {
      icon: Users,
      value: "0",
      label: t("projects.members"),
      color: "text-blue-500",
      bg: "bg-blue-500/10",
    },
    {
      icon: FileText,
      value: "0",
      label: t("projects.documents"),
      color: "text-emerald-500",
      bg: "bg-emerald-500/10",
    },
    {
      icon: BarChart3,
      value: "0",
      label: t("projects.estimations"),
      color: "text-amber-500",
      bg: "bg-amber-500/10",
    },
  ];

  const statusColor = (status: string) => {
    switch (status) {
      case "active":
        return "text-emerald-500";
      case "archived":
        return "text-muted-foreground";
      default:
        return "text-blue-500";
    }
  };

  const statusLabel = (status: string) => {
    switch (status) {
      case "active":
        return t("projects.active");
      case "archived":
        return t("projects.archived");
      default:
        return status;
    }
  };

  return (
    <div>
      {/* Welcome */}
      <div className="mb-10">
        <div className="flex items-center gap-2 mb-1">
          <Sparkles className="h-5 w-5 text-amber-500" />
          <h1 className="text-2xl font-bold tracking-tight">{greeting}</h1>
        </div>
        <p className="text-muted-foreground">{t("dashboard.welcomeSubtitle")}</p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-12">
        {stats.map((stat) => (
          <div
            key={stat.label}
            className="rounded-xl p-5 text-center flex flex-col items-center"
          >
            <div className={`inline-flex items-center justify-center h-9 w-9 rounded-lg ${stat.bg} mb-3`}>
              <stat.icon className={`h-5 w-5 ${stat.color}`} />
            </div>
            <p className="text-2xl font-bold">{stat.value}</p>
            <p className="text-sm text-muted-foreground">{stat.label}</p>
          </div>
        ))}
      </div>

      {/* Empty state or project cards */}
      {projects.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 mb-12">
          <div className="relative mb-6">
            <div className="absolute inset-0 rounded-full bg-gradient-to-br from-violet-500/20 to-blue-500/20 blur-xl" />
            <div className="relative h-16 w-16 rounded-full bg-muted/50 flex items-center justify-center">
              <FolderKanban className="h-7 w-7 text-muted-foreground" />
            </div>
          </div>
          <p className="text-lg font-medium mb-1">{t("projects.empty")}</p>
          <p className="text-sm text-muted-foreground mb-8 text-center max-w-sm">
            {t("projects.emptyDesc")}
          </p>
          <button
            className="inline-flex items-center gap-2 rounded-lg border border-border bg-transparent px-6 py-3 text-sm font-medium text-foreground transition-colors hover:bg-muted hover:text-muted-foreground"
            onClick={() => setCreateOpen(true)}
          >
            <Plus className="h-4 w-4" />
            {t("projects.create")}
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {projects.map((project) => (
            <Link
              key={project.id}
              href={`/dashboard/projects/${project.id}`}
              className="rounded-2xl border border-foreground/10 bg-background p-6 cursor-pointer transition-colors duration-200 hover:bg-muted block"
            >
              <div className="inline-flex items-center justify-center h-10 w-10 rounded-lg bg-muted/50 mb-4">
                <FolderOpen className="h-5 w-5 text-foreground" />
              </div>
              <h3 className="text-lg font-semibold mb-2">{project.name}</h3>
              <p className="text-sm text-muted-foreground mb-4 line-clamp-2">
                {project.description || "\u00A0"}
              </p>
              <span className={`text-sm font-medium ${statusColor(project.status)}`}>
                {statusLabel(project.status)}
              </span>
            </Link>
          ))}

          {/* Add project card */}
          <div
            className="relative rounded-2xl cursor-pointer group"
            onClick={() => setCreateOpen(true)}
          >
            <div className="relative rounded-2xl border border-dashed border-foreground/20 bg-background p-6 h-full flex flex-col items-center justify-center transition-colors duration-300 hover:border-foreground/40">
              <div className="inline-flex items-center justify-center h-10 w-10 rounded-lg bg-muted/50 mb-4">
                <Plus className="h-5 w-5 text-muted-foreground" />
              </div>
              <p className="text-sm text-muted-foreground">{t("projects.create")}</p>
            </div>
          </div>
        </div>
      )}

      {/* Create project dialog */}
      <CreateProjectDialog
        open={createOpen}
        onOpenChange={setCreateOpen}
        workspaceId={workspaceId ?? ""}
      />
    </div>
  );
}
