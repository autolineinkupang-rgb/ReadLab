const API_BASE = "/api";

function getCSRFToken(): string {
  const match = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : "";
}

async function fetcherFormData<T>(endpoint: string, formData: FormData): Promise<T> {
  const csrfToken = getCSRFToken();
  const headers: Record<string, string> = {};
  if (csrfToken) headers["X-CSRF-Token"] = csrfToken;

  const res = await fetch(`${API_BASE}${endpoint}`, {
    method: "POST",
    credentials: "include",
    headers,
    body: formData,
  });

  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(error.error || `HTTP ${res.status}`);
  }

  return res.json();
}

interface FetcherOptions extends RequestInit {
  params?: Record<string, string | number | undefined>;
}

async function fetcher<T>(endpoint: string, options: FetcherOptions = {}): Promise<T> {
  const { params, ...init } = options;

  let url = `${API_BASE}${endpoint}`;

  if (params) {
    const searchParams = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== "") {
        searchParams.set(key, String(value));
      }
    });
    const qs = searchParams.toString();
    if (qs) url += `?${qs}`;
  }

  const method = (init.method || "GET").toUpperCase();
  const isMutation = method === "POST" || method === "PUT" || method === "PATCH" || method === "DELETE";
  const csrfToken = isMutation ? getCSRFToken() : "";

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init.headers as Record<string, string> | undefined),
  };

  if (isMutation && csrfToken) {
    headers["X-CSRF-Token"] = csrfToken;
  }

  const res = await fetch(url, {
    ...init,
    credentials: "include",
    headers,
  });

  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(error.error || `HTTP ${res.status}`);
  }

  return res.json();
}

// Auth
export const auth = {
  login: (email: string, password: string) =>
    fetcher<{ user: any }>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),
  register: (username: string, email: string, password: string) =>
    fetcher<{ user: any }>("/auth/register", {
      method: "POST",
      body: JSON.stringify({ username, email, password }),
    }),
  me: () => fetcher<any>("/auth/me"),
  logout: () => fetcher<{ message: string }>("/auth/logout", { method: "POST" }),
  changePassword: (currentPassword: string, newPassword: string) =>
    fetcher<{ message: string }>("/auth/password", {
      method: "PUT",
      body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
    }),
  forgotPassword: (email: string) =>
    fetcher<{ message: string; reset_link?: string }>("/auth/forgot-password", {
      method: "POST",
      body: JSON.stringify({ email }),
    }),
  resetPassword: (token: string, newPassword: string) =>
    fetcher<{ message: string }>("/auth/reset-password", {
      method: "POST",
      body: JSON.stringify({ token, new_password: newPassword }),
    }),
};

// Rewards
export const rewards = {
  daily: () => fetcher<{ message: string; tickets: number; rewarded: number }>("/rewards/daily", { method: "POST" }),
  status: () => fetcher<{ daily_reward: { can_claim: boolean; reward: number; next_claim_at?: string } }>("/rewards/status"),
};

export const xpConfig = {
  list: () => fetcher<Record<string, number>>("/config/xp"),
};

// Novels
export const novels = {
  list: (params?: { page?: number; limit?: number; q?: string; status?: string; genre?: string; genres?: string; genre_mode?: string; sort?: string; order?: string; min_chapters?: number; min_rating?: number; min_reviews?: number; writer_id?: number }) =>
    fetcher<{ data: any[]; page: number; limit: number; total: number; total_pages: number }>("/novels", { params: params as any }),
  get: (id: number | string) => fetcher<any>(`/novels/${id}`),
  chapters: (id: number | string, params?: { page?: number; limit?: number }) =>
    fetcher<{ data: any[]; page: number; limit: number; total: number }>(`/novels/${id}/chapters`, { params: params as any }),
  trending: () => fetcher<{ data: any[] }>("/novels/trending"),
  recommendations: () => fetcher<{ data: any[] }>("/novels/recommendations"),
  random: (limit?: number) => fetcher<{ data: any[] }>("/novels/random", { params: { limit } }),
};

// Chapters
export const chapters = {
  get: (id: number | string) => fetcher<any>(`/chapters/${id}`),
  getByNovel: (novelId: number | string, num: number) =>
    fetcher<any>(`/novels/${novelId}/chapters/${num}`),
  importMd: (novelId: number | string, file: File, mode: "preview" | "save") => {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("mode", mode);
    return fetcherFormData<any>(`/writer/novels/${novelId}/chapters/import-md`, formData);
  },
};

// Notifications
export const notifications = {
  list: () => fetcher<{ data: any[]; unread_count: number }>("/notifications"),
  unreadCount: () => fetcher<{ unread_count: number }>("/notifications/unread-count"),
  markRead: (id: number | "all") => fetcher<{ message: string }>(`/notifications/${id}/read`, { method: "PUT" }),
};

// Tickets / Purchase
export const tickets = {
  purchase: (amount: number) =>
    fetcher<{ message: string; amount: number }>("/tickets/purchase", {
      method: "POST",
      body: JSON.stringify({ amount }),
    }),
};

// Ranking
export const ranking = {
  get: (period: string = "daily") => fetcher<{ data: any[] }>(`/ranking/${period}`),
};

// Updates
export const updates = {
  recent: (limit?: number) => fetcher<{ data: any[] }>("/updates", { params: { limit } }),
};

// Search
export const search = {
  query: (q: string, params?: { page?: number; limit?: number }) =>
    fetcher<{ data: any[]; page: number; limit: number; total: number }>("/search", { params: { q, ...params } as any }),
  autocomplete: (q: string) =>
    fetcher<{ data: { id: number; slug: string; title: string }[] }>("/search/autocomplete", { params: { q } as any }),
};

// Genres
export const genres = {
  list: () => fetcher<{ data: any[] }>("/genres"),
};

// Tags
export const tagsApi = {
  list: () => fetcher<{ data: { ID: number; Name: string; Slug: string }[] }>("/tags"),
};

// Config
export const upgradeCosts = () =>
  fetcher<{ edit_reset: number; gate_bypass: number; replace_review: number }>("/config/upgrade-costs");

// Leaderboard
export const leaderboard = {
  get: (sort?: string) => fetcher<{ data: any[] }>("/leaderboard", { params: { sort } }),
};

// News
export const news = {
  list: (params?: { type?: string; page?: number; limit?: number }) =>
    fetcher<{ data: any[]; page: number; limit: number; total: number }>("/news", { params: params as any }),
  get: (id: number | string) => fetcher<any>(`/news/${id}`),
};

// Votes (protected)
export const votes = {
  create: (novelId: number) =>
    fetcher<{ message: string; xp_earned: number }>("/votes", {
      method: "POST",
      body: JSON.stringify({ novel_id: novelId }),
    }),
};

// Requests (protected)
export const requests = {
  list: () =>
    fetcher<{ data: any[] }>("/requests"),
  create: (data: { novel_title: string; novel_url?: string; source?: string }) =>
    fetcher<any>("/requests", {
      method: "POST",
      body: JSON.stringify(data),
    }),
};

// Library (protected)
export const library = {
  get: () => fetcher<{ follows: any[]; history: any[] }>("/library"),
};

// Follow (protected)
export const follow = {
  create: (novelId: number) =>
    fetcher<{ message: string }>(`/novels/${novelId}/follow`, { method: "POST" }),
  delete: (novelId: number) =>
    fetcher<{ message: string }>(`/novels/${novelId}/follow`, { method: "DELETE" }),
  check: (novelId: number) =>
    fetcher<{ following: boolean }>(`/novels/${novelId}/follow`),
};

// Author
export const author = {
  novels: (name: string) => fetcher<{ data: any[]; total: number }>(`/author/${encodeURIComponent(name)}/novels`),
};

// Profile
export const profile = {
  get: (id: string | number) => fetcher<any>(`/profile/${id}`),
};

// Stats
export const stats = {
  get: () => fetcher<{
    total_novels: number;
    total_chapters: number;
    total_users: number;
    total_views: number;
    total_votes: number;
    total_requests: number;
  }>("/stats"),
};

// Health
export const health = {
	check: () => fetcher<{ status: string }>("/health"),
};

// Import (protected)
export const importer = {
	search: (q: string) => fetcher<{ data: { id: string; title: string; url: string; image: string }[] }>(`/novels/import/search?q=${encodeURIComponent(q)}`),
	import: (sourceID: string, withChapters?: boolean) =>
		fetcher<{ data: any }>("/novels/import", {
			method: "POST",
			body: JSON.stringify({ source_id: sourceID, with_chapters: withChapters ?? false }),
		}),
};

// Lncrawl (protected)
export const lncrawl = {
	crawl: (url: string, maxChapters?: number, chapterRange?: string) =>
		fetcher<{ data: any }>("/novels/lncrawl", {
			method: "POST",
			body: JSON.stringify({ url, max_chapters: maxChapters ?? 0, chapter_range: chapterRange }),
		}),
};

// Web Scraper (protected)
export const scraper = {
	scrape: (url: string) =>
		fetcher<{ data: any }>("/novels/scrape", {
			method: "POST",
			body: JSON.stringify({ url }),
		}),
	import: (url: string, withContent: boolean, chapterRange?: string) =>
		fetcher<{ data: any }>("/novels/scrape/import", {
			method: "POST",
			body: JSON.stringify({ url, with_content: withContent, chapter_range: chapterRange }),
		}),
};

// Writer chapters (protected)
export const writerChapters = {
  list: (novelId: number | string, params?: { page?: number; limit?: number }) =>
    fetcher<{ data: any[]; total: number; page: number; limit: number; total_pages: number }>(`/writer/novels/${novelId}/chapters`, { params: params as any }),
  get: (id: number | string) =>
    fetcher<{ chapter: any }>(`/writer/chapters/${id}`),
  create: (novelId: number | string, data: { number?: number; title: string; content: string; is_locked?: boolean; ticket_cost?: number }) =>
    fetcher<{ chapter: any }>(`/writer/novels/${novelId}/chapters`, {
      method: "POST",
      body: JSON.stringify(data),
    }),
  update: (novelId: number | string, chapterId: number | string, data: { number?: number; title?: string; content?: string; is_locked?: boolean; ticket_cost?: number }) =>
    fetcher<{ chapter: any }>(`/writer/novels/${novelId}/chapters/${chapterId}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),
  delete: (id: number | string) =>
    fetcher<{ message: string }>(`/writer/chapters/${id}`, {
      method: "DELETE",
    }),
};

// Admin novels (protected)
export const adminNovels = {
	create: (data: any) =>
		fetcher<{ data: any }>("/novels", {
			method: "POST",
			body: JSON.stringify(data),
		}),
	update: (id: number | string, data: any) =>
		fetcher<{ data: any }>(`/novels/${id}`, {
			method: "PUT",
			body: JSON.stringify(data),
		}),
	delete: (id: number | string) =>
		fetcher<{ message: string }>(`/novels/${id}`, {
			method: "DELETE",
		}),
};

// Translation
export const translateApi = {
	translate: (text: string, target: string, source?: string) =>
		fetcher<{ data: string }>("/translate", {
			method: "POST",
			body: JSON.stringify({ text, target, source }),
		}),
};

// AI Translation settings
export interface AITranslateSettings {
	provider: string;
	model: string;
	endpoint: string;
	key: string;
	has_key: boolean;
	target_language: string;
	instruction: string;
}

export const aiSettings = {
	get: () => fetcher<AITranslateSettings>("/user/ai-settings"),
	update: (data: { provider: string; model: string; endpoint: string; key?: string; target_language?: string; instruction?: string }) =>
		fetcher<{ message: string }>("/user/ai-settings", {
			method: "PUT",
			body: JSON.stringify(data),
		}),
};

export const translateAi = {
	translate: (text: string, target?: string, source?: string) =>
		fetcher<{ data: string }>("/translate/ai", {
			method: "POST",
			body: JSON.stringify({ text, target, source }),
		}),
};

// Admin requests (protected)
export const adminRequests = {
	list: (params?: { page?: number; limit?: number; status?: string }) =>
		fetcher<{ data: any[]; total: number; page: number; limit: number; total_pages: number }>("/admin/requests", { params: params as any }),
	review: (id: number | string, status: string) =>
		fetcher<any>(`/requests/${id}`, {
			method: "PUT",
			body: JSON.stringify({ status }),
		}),
};

// Admin reviews (protected)
export const adminReviews = {
	list: (params?: { page?: number; limit?: number }) =>
		fetcher<{ data: any[]; total: number; page: number; limit: number; total_pages: number }>("/admin/reviews", { params: params as any }),
	delete: (id: number | string) =>
		fetcher<{ message: string }>("/admin/reviews/" + id, { method: "DELETE" }),
};

// Admin news (protected)
export const adminNews = {
  create: (data: { title: string; content: string; type: string }) =>
    fetcher<any>("/admin/news", { method: "POST", body: JSON.stringify(data) }),
  update: (id: number | string, data: { title?: string; content?: string; type?: string }) =>
    fetcher<any>("/admin/news/" + id, { method: "PUT", body: JSON.stringify(data) }),
  delete: (id: number | string) =>
    fetcher<{ message: string }>("/admin/news/" + id, { method: "DELETE" }),
};

// Admin ticket config
export const adminTicketConfig = {
  list: () => fetcher<{ data: { id: number; key: string; value: number; label: string }[] }>("/admin/config/tickets"),
  update: (key: string, value: number) =>
    fetcher<{ message: string }>("/admin/config/tickets", {
      method: "PUT",
      body: JSON.stringify({ key, value }),
    }),
};

export const adminXpConfig = {
  list: () => fetcher<Record<string, number>>("/config/xp"),
  update: (key: string, value: number) =>
    fetcher<{ message: string }>("/admin/config/tickets", {
      method: "PUT",
      body: JSON.stringify({ key, value }),
    }),
};

// Admin users (protected)
export const adminUsers = {
	list: (params?: { page?: number; limit?: number; role?: string; q?: string }) =>
		fetcher<{ data: any[]; page: number; limit: number; total: number; total_pages: number }>("/admin/users", { params: params as any }),
	get: (id: number | string) => fetcher<any>(`/admin/users/${id}`),
	update: (id: number | string, data: { role?: string; tickets?: number }) =>
		fetcher<any>(`/admin/users/${id}`, {
			method: "PUT",
			body: JSON.stringify(data),
		}),
	delete: (id: number | string) =>
		fetcher<any>(`/admin/users/${id}`, {
			method: "DELETE",
		}),
	sendTickets: (id: number | string, amount: number) =>
		fetcher<any>(`/admin/users/${id}/tickets`, {
			method: "POST",
			body: JSON.stringify({ amount }),
		}),
	createAdmin: (data: { username: string; email: string; password: string }) =>
		fetcher<any>("/admin/users/admin", {
			method: "POST",
			body: JSON.stringify(data),
		}),
	stats: () =>
		fetcher<{ total_users: number; total_novels: number; total_chapters: number; total_admins: number; max_admins: number }>("/admin/stats"),
};

export const adminBank = {
	balance: () => fetcher<{ balance: number; units: number }>("/admin/bank"),
	claim: (amount: number) =>
		fetcher<{ message: string }>("/admin/bank/claim", {
			method: "POST",
			body: JSON.stringify({ amount }),
		}),
};

export interface ApiError extends Error {
  upgrade_available?: boolean;
  upgrade_cost?: number;
  upgrade_type?: string;
}

async function reviewFetcher<T>(endpoint: string, options: RequestInit & { params?: Record<string, string | number | undefined> }): Promise<T> {
  const { params, ...init } = options;
  let url = `/api${endpoint}`;
  const method = (init.method || "GET").toUpperCase();
  const isMutation = method === "POST" || method === "PUT" || method === "PATCH" || method === "DELETE";
  const csrfToken = isMutation ? getCSRFToken() : "";
  const headers: Record<string, string> = { "Content-Type": "application/json", ...init.headers as Record<string, string> };
  if (isMutation && csrfToken) {
    headers["X-CSRF-Token"] = csrfToken;
  }
  const res = await fetch(url, {
    ...init,
    credentials: "include",
    headers,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    const err = new Error(body.error || `HTTP ${res.status}`) as ApiError;
    if (body.upgrade_available) {
      err.upgrade_available = true;
      err.upgrade_cost = body.upgrade_cost;
      err.upgrade_type = body.upgrade_type;
    }
    throw err;
  }
  return res.json();
}

// Reviews
export const reviews = {
  list: (novelId: number) =>
    fetcher<{ data: ReviewResponse[]; rating_summary: RatingSummary }>(`/novels/${novelId}/reviews`),
  create: (novelId: number, rating: number, content: string, upgrade?: boolean, parent_id?: number) =>
    reviewFetcher<{ data: ReviewResponse; xp_earned?: number }>(`/novels/${novelId}/reviews`, {
      method: "POST",
      body: JSON.stringify({ rating, content, parent_id, upgrade }),
    }),
  update: (novelId: number, reviewId: number, rating: number, content: string, upgrade?: boolean) =>
    reviewFetcher<{ data: ReviewResponse }>(`/novels/${novelId}/reviews/${reviewId}`, {
      method: "PUT",
      body: JSON.stringify({ rating, content, upgrade }),
    }),
};

// Shares
export const shares = {
  create: (novelId: number, platform: string) =>
    fetcher<{ message: string; xp_earned: number }>(`/novels/${novelId}/share`, {
      method: "POST",
      body: JSON.stringify({ platform }),
    }),
};

// Reading tracking
export const reading = {
  track: (novelId: number, chapterNum: number) =>
    fetcher<{ message: string }>(`/novels/${novelId}/chapters/${chapterNum}/read`, {
      method: "POST",
    }),
  claimXP: (novelId: number, chapterNum: number) =>
    fetcher<{ message: string; xp_earned: number }>(`/novels/${novelId}/chapters/${chapterNum}/xp`, {
      method: "POST",
    }),
  progress: (novelId: number) =>
    fetcher<{ chapter_count: number; can_review: boolean; my_review: ReviewResponse | null; last_chapter?: number; last_chapter_title?: string }>(
      `/novels/${novelId}/my-progress`
    ),
};

// Types for reviews
export interface ReviewResponse {
  id: number;
  rating: number;
  content: string;
  edit_count: number;
  parent_id: number | null;
  created_at: string;
  user: {
    id: number;
    username: string;
    display_name: string;
    avatar_url: string;
  };
  replies: ReviewResponse[];
}

export interface RatingSummary {
  average: number;
  count: number;
  distribution: Record<number, number>;
}
