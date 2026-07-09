export interface Genre {
  ID?: number;
  Slug: string;
  Name: string;
}

export interface Novel {
  ID: number;
  Title: string;
  AltTitle?: string;
  Slug: string;
  Author?: string;
  AuthorSlug?: string;
  Status?: string;
  Views: number;
  Rating: number;
  RatingCount?: number;
  Chapters: number;
  Readers?: number;
  Chars?: string;
  AIPercent?: string;
  Description?: string;
  CoverURL: string;
  RequestedBy?: string;
  ReleasedBy?: string;
  Genres: Genre[];
  CreatedAt?: string;
  Tags?: { ID: number; Name: string; Slug: string }[];
  ReleaseStatus?: string;
  AddedMinutesAgo?: number;
  Votes?: number;
}

export interface Chapter {
  ID: number;
  NovelID: number;
  Number: number;
  Title: string;
  IsLocked: boolean;
  TicketCost: number;
  Content?: string;
  CreatedAt?: string;
}

export interface ProfileData {
  id: number;
  username: string;
  display_name: string;
  avatar_url: string;
  tickets: number;
  xp: number;
  created_at: string;
}

