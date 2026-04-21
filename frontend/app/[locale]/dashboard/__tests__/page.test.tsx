import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useAuthStore } from "@/features/auth/store";

vi.mock("@/lib/query-client", () => ({
  getQueryClient: () => new QueryClient({ defaultOptions: { queries: { retry: false } } }),
}));

import DashboardPage from "../page";

function renderWithProviders() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <DashboardPage />
    </QueryClientProvider>,
  );
}

beforeEach(() => {
  localStorage.setItem("ep_access_token", "test-token");
  useAuthStore.setState({
    user: { id: "u1", name: "Test User", email: "test@test.com", preferred_locale: "ru" },
    isAuthenticated: true,
    isLoading: false,
  });
});

describe("DashboardPage", () => {
  it("renders welcome greeting", () => {
    renderWithProviders();
    expect(screen.getByText(/dashboard.welcomeBack/)).toBeInTheDocument();
  });

  it("renders stat cards section", async () => {
    renderWithProviders();
    expect(await screen.findByText("projects.title")).toBeInTheDocument();
  });

  it("renders workspace cards from API", async () => {
    renderWithProviders();
    // MSW returns workspace "Default Workspace"
    expect(await screen.findByText("Default Workspace")).toBeInTheDocument();
  });

  it("renders project data from API", async () => {
    renderWithProviders();
    // MSW returns project "Project 1"
    expect(await screen.findByText("Project 1")).toBeInTheDocument();
  });

  it("renders multiple stat sections", async () => {
    renderWithProviders();
    // Wait for data to load, then check multiple sections render
    await screen.findByText("projects.title");
    const headings = screen.getAllByRole("heading");
    expect(headings.length).toBeGreaterThanOrEqual(1);
  });
});
