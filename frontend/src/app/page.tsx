import Link from "next/link";
import NovelCard from "@/components/NovelCard";
import SectionHeader from "@/components/SectionHeader";
import NovelCardSmall from "@/components/NovelCardSmall";
import UpdateItem from "@/components/UpdateItem";

const newNovels = [
  { title: "Having Dinner with His Brother, the Cold and Aloof Tycoon Becomes Addicted to His Doting Affections", genre: "romance", chapters: 135, href: "/novel/1" },
  { title: "Corpse Puppet Phoenix Girl", genre: "adult", chapters: 242, href: "/novel/2" },
  { title: "Reborn As the Little Delicate Wife of the Domineering Ceo", genre: "romance", chapters: 378, href: "/novel/3" },
  { title: "The Corpse Family is Heavy", genre: "action", chapters: 252, href: "/novel/4" },
  { title: "Can You Please Comfort Me?", genre: "drama", chapters: 149, href: "/novel/5" },
  { title: "First-rank Di Consort", genre: "historical", chapters: 387, href: "/novel/6" },
  { title: "I am the Crown Prince of the Ming Dynasty", genre: "action", chapters: 1592, href: "/novel/7" },
  { title: "The Legend of the Mountain and Sea Demon Subduing", genre: "action", chapters: 1522, href: "/novel/8" },
  { title: "Don't Be Too Wild", genre: "romance", chapters: 160, href: "/novel/9" },
  { title: "The Villious Prince is Three", genre: "adult", chapters: 281, href: "/novel/10" },
];

const rankingNovels = [
  { rank: 1, title: "Question and Answer Douluo: Tang San's Time Travel Revealed, Tang Hao Breaks Through Defense", views: "167,204", rating: "1.7", href: "/novel/11" },
  { rank: 2, title: "Douluo Continent: Taking Tang San As a Disciple, with a Ten-thousand-fold Return for Teaching Him", views: "61,300", rating: "2.0", href: "/novel/12" },
  { rank: 3, title: "Lord: God-tier Attribute, Recruits Fallen Angels of Original Sin", views: "50,895", rating: "3.3", href: "/novel/13" },
  { rank: 4, title: "I Just Started High School, But the System Insists I'm an Emperor in My Twilight Years", views: "50,315", rating: "1.9", href: "/novel/14" },
  { rank: 5, title: "Football: I See Weaknesses!", views: "41,170", rating: "1.8", href: "/novel/15" },
];

const trendingCovers = [
  "Naruto: In Konoha Village, I Awakened Wood Release at the Start",
  "Question and Answer Douluo: Tang San's Time Travel Revealed, Tang Hao Breaks Through Defense",
  "Hogwarts: My Magic Has Turned Evil!",
  "Battle Through the Heavens: I Can Solidify the Talents of All Things",
  "Douluo Continent: A Fictional Sky, Turning Fiction Into Reality",
  "As Long As I Lack Morality, Konoha Can't Do Anything to Me!",
  "Football: I See Weaknesses!",
  "Marvel: Checking in at New York, Starting with a Silver Superman",
  "Douluo Continent: Taking Tang San As a Disciple, with a Ten-thousand-fold Return for Teaching Him",
  "Primordial Era: Expelled from the Jie Sect? I'll Open the Floodgates, What Will You Regret Then?",
  "Douluo Continent: Qian Renxue, Your Martial Soul Has Become Sentient!",
  "The Cruelty of the Uchiha",
];

const recommendations = [
  { title: "Question and Answer Douluo: Tang San's Time Travel Revealed, Tang Hao Breaks Through Defense", genre: "action", rating: "1.7", chapters: 372, href: "/novel/11" },
  { title: "Douluo Continent: Taking Tang San As a Disciple, with a Ten-thousand-fold Return for Teaching Him", genre: "action", rating: "2.0", chapters: 284, href: "/novel/12" },
  { title: "I Just Started High School, But the System Insists I'm an Emperor in My Twilight Years", genre: "fantasy", rating: "1.9", chapters: 156, href: "/novel/14" },
  { title: "All Heavens: My Dantian is a Universe in Its Own", genre: "action", rating: "3.2", chapters: 412, href: "/novel/16" },
  { title: "Pretending to Be the Villain, All I Want is to Die at Naruto's Hands", genre: "action", rating: "2.8", chapters: 98, href: "/novel/17" },
  { title: "On the Day the Empress Sentenced Me to Death, the System Granted Me Emperor Level Cultivation", genre: "fantasy", rating: "3.5", chapters: 203, href: "/novel/18" },
];

const recentUpdates = [
  { title: "Reborn in 2002, I'm Carrying a Huawei Phone", chapter: "#184 Chapter 184: 4,000 Units Sold in the First Phase", chapterHref: "/novel/19/chapter-184", novelHref: "/novel/19", hasImage: false },
  { title: "Only After Ascending to Heaven Did I Realize That This Was Journey to the West", chapter: "#190 Chapter 190: Each Goes Their Own Way!", chapterHref: "/novel/20/chapter-190", novelHref: "/novel/20", hasImage: true },
  { title: "Douluo: Starting with the Role of a Born God of Strength", chapter: "#80 Chapter 80: False Accusation!", chapterHref: "/novel/21/chapter-80", novelHref: "/novel/21", hasImage: true },
  { title: "World War II Began with Preparations in Britain", chapter: "#102 Chapter 102: The Power of Financial Magic", chapterHref: "/novel/22/chapter-102", novelHref: "/novel/22", hasImage: true },
  { title: "From Dandelion to Evolution Into Dandelion Tree", chapter: "#254 Chapter 254: One in Ten Thousand Power", chapterHref: "/novel/23/chapter-254", novelHref: "/novel/23", hasImage: true },
  { title: "Upscale Residential Area Ahead, No Impostors Allowed", chapter: "#311 Chapter 311: The Immortal Under the Mask", chapterHref: "/novel/24/chapter-311", novelHref: "/novel/24", hasImage: true },
];

const randomNovels = [
  { title: "Binding the Shanhaijing Pearl at the Beginning, I Became the Global Treasure Hunt King", genre: "action", chapters: 122, rating: "3.1", href: "/novel/25" },
  { title: "Peninsula: Kpop Hit Maker", genre: "drama", chapters: 191, rating: "2.2", href: "/novel/26" },
  { title: "Wuxia Crossover: Sweeping the Heavens, Fun Fun Fun Fun Fun Fun Fun", genre: "action", chapters: 501, rating: "2.5", href: "/novel/27" },
  { title: "Invincible Heavenly Emperor", genre: "action", chapters: 3871, rating: "2.5", href: "/novel/28" },
  { title: "Naruto: I Was Spoiled by the Heavenly Curtain to Unify the Ninja World", genre: "action", chapters: 96, rating: "2.5", href: "/novel/29" },
  { title: "The Marvel World of Heroes", genre: "action", chapters: 469, rating: "2.5", href: "/novel/30" },
  { title: "Start with Uchiha to escape and sail", genre: "action", chapters: 434, rating: "3.3", href: "/novel/31" },
  { title: "There are No Ancestors. They are All Made Up by Me.", genre: "action", chapters: 328, rating: "3.4", href: "/novel/32" },
  { title: "Overthrow the Han Dynasty", genre: "action", chapters: 289, rating: "1.0", href: "/novel/33" },
  { title: "A Crossover Anime/manga Business, Starting with the Ten Holy Blades Saving Himeko", genre: "action", chapters: 158, rating: "4.1", href: "/novel/34" },
];

const topSpenders = [
  { name: "Mega_bells", tickets: "3,569.76", href: "/profile/1" },
  { name: "StandardCrystal", tickets: "2,907.17", href: "/profile/2" },
  { name: "Alpha2", tickets: "2,693.07", href: "/profile/3" },
];

export default function Home() {
  return (
    <div className="max-w-7xl mx-auto px-4 py-6 space-y-10">
      {/* Giveaway Banner */}
      <div className="bg-gradient-to-r from-violet-900/40 to-purple-900/40 border border-violet-800/30 rounded-xl p-6 text-center">
        <h3 className="text-xl font-bold text-white mb-2">🎉 16th Giveaway Winners 🎉</h3>
        <Link href="/news/428" className="inline-block mt-2 px-4 py-2 bg-violet-600 hover:bg-violet-700 text-white text-sm rounded-lg transition-colors">
          Check Results
        </Link>
      </div>

      {/* Login Prompt */}
      <div className="text-center text-sm text-gray-500">
        Login to keep track of where you left off in the novel.
      </div>

      {/* New Novels */}
      <section>
        <SectionHeader title="New Novels" href="/novel-list" />
        <div className="flex gap-4 overflow-x-auto pb-2 scrollbar-hide">
          {newNovels.map((novel) => (
            <NovelCard key={novel.href} {...novel} />
          ))}
        </div>
      </section>

      {/* Novel Ranking + Trending Row */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-8">
        {/* Ranking - takes 2 cols */}
        <div className="lg:col-span-2">
          <SectionHeader title="Novel Ranking" href="/ranking/daily" tabs={[{ label: "Daily", active: true }, { label: "Weekly" }, { label: "Monthly" }]} />
          <div className="bg-[#12122a] border border-[#1e1e3a] rounded-xl p-4 space-y-1">
            {rankingNovels.map((novel) => (
              <NovelCardSmall key={novel.rank} {...novel} />
            ))}
          </div>
          <div className="flex gap-3 mt-4">
            <Link href="/profile/vote-serie" className="flex-1 text-center text-sm py-2.5 rounded-lg bg-[#1e1e3a] hover:bg-[#2a2a4a] text-gray-300 transition-colors">
              Vote Novels
            </Link>
            <Link href="/profile/request-serie" className="flex-1 text-center text-sm py-2.5 rounded-lg bg-[#1e1e3a] hover:bg-[#2a2a4a] text-gray-300 transition-colors">
              Request Novels
            </Link>
          </div>
        </div>

        {/* Trending - takes 2 cols */}
        <div className="lg:col-span-2">
          <SectionHeader title="Trending" href="/trending" />
          <div className="flex gap-3 overflow-x-auto pb-2 scrollbar-hide">
            {trendingCovers.map((title, i) => (
              <div key={i} className="w-24 flex-shrink-0">
                <div className="aspect-[3/4] rounded-lg bg-[#1e1e3a] border border-[#2a2a4a] flex items-center justify-center">
                  <svg className="w-8 h-8 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                  </svg>
                </div>
                <p className="text-[10px] text-gray-400 mt-1 line-clamp-2 leading-tight">{title}</p>
              </div>
            ))}
          </div>

          {/* Featured Novel */}
          <div className="mt-6 bg-[#12122a] border border-[#1e1e3a] rounded-xl p-5">
            <div className="flex gap-4">
              <div className="w-24 sm:w-32 aspect-[3/4] rounded-lg bg-[#1e1e3a] border border-[#2a2a4a] flex-shrink-0 flex items-center justify-center">
                <svg className="w-10 h-10 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
                </svg>
              </div>
              <div className="min-w-0 flex-1">
                <Link href="/novel/featured" className="text-base font-semibold text-white hover:text-violet-400 transition-colors line-clamp-2">
                  Naruto: In Konoha Village, I Awakened Wood Release at the Start
                </Link>
                <div className="flex items-center gap-3 mt-1 text-sm text-gray-400">
                  <span>★ 3.8</span>
                  <span>📚 1003</span>
                </div>
                <div className="flex flex-wrap gap-1.5 mt-2">
                  {["action", "fan-fiction", "fantasy", "martial-arts", "seinen", "supernatural"].map((tag) => (
                    <span key={tag} className="text-xs px-2 py-0.5 rounded-full bg-violet-900/40 text-violet-300 border border-violet-800/30">
                      {tag}
                    </span>
                  ))}
                </div>
                <p className="text-sm text-gray-400 mt-2 line-clamp-3 leading-relaxed">
                  Konoha 52nd year. Chiba awakened her memories of her past life. As one of the descendants of the Senju clan. With the help of the system, Chiba successfully awakened Wood Release. But the first thing he did was to find a way to escape from Konoha Village.
                </p>
                <Link
                  href="/novel/featured/continue"
                  className="inline-block mt-3 px-4 py-1.5 bg-violet-600 hover:bg-violet-700 text-white text-sm rounded-lg transition-colors"
                >
                  START READING
                </Link>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Recommendations */}
      <section>
        <SectionHeader title="Recommendations" href="/recommendation" />
        <div className="flex gap-4 overflow-x-auto pb-2 scrollbar-hide">
          {recommendations.map((novel, i) => (
            <NovelCard key={i} {...novel} compact />
          ))}
        </div>
        {/* Featured Recommendation */}
        <div className="mt-4 bg-[#12122a] border border-[#1e1e3a] rounded-xl p-5">
          <div className="flex gap-4">
            <div className="w-20 sm:w-28 aspect-[3/4] rounded-lg bg-[#1e1e3a] border border-[#2a2a4a] flex-shrink-0 flex items-center justify-center">
              <svg className="w-8 h-8 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253" />
              </svg>
            </div>
            <div className="min-w-0 flex-1">
              <Link href="/novel/featured-rec" className="text-base font-semibold text-white hover:text-violet-400 transition-colors line-clamp-2">
                Question and Answer Douluo: Tang San&apos;s Time Travel Revealed, Tang Hao Breaks Through Defense
              </Link>
              <div className="flex items-center gap-3 mt-1 text-sm text-gray-400">
                <span>★ 1.7</span>
                <span>📚 372</span>
              </div>
              <div className="flex flex-wrap gap-1.5 mt-2">
                {["action", "adventure", "fan-fiction", "fantasy"].map((tag) => (
                  <span key={tag} className="text-xs px-2 py-0.5 rounded-full bg-violet-900/40 text-violet-300 border border-violet-800/30">
                    {tag}
                  </span>
                ))}
              </div>
              <p className="text-sm text-gray-400 mt-2 line-clamp-3 leading-relaxed">
                Transmigrating into Douluo Continent, Song Ye becomes a member of the Spirit Hall team. During the Soul Master Competition, a [Douluo Quiz Game] suddenly appears! Answer correctly to receive unlimited rewards! Song Ye, having thoroughly studied the original novel, keeps racking up rewards...
              </p>
              <Link
                href="/novel/featured-rec/continue"
                className="inline-block mt-3 px-4 py-1.5 bg-violet-600 hover:bg-violet-700 text-white text-sm rounded-lg transition-colors"
              >
                START READING
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Bug Reports / Patreon */}
      <div className="flex flex-wrap gap-4 justify-center">
        <Link href="https://discord.gg/wtrlab" className="flex items-center gap-2 px-5 py-2.5 bg-[#1e1e3a] hover:bg-[#2a2a4a] rounded-lg text-sm text-gray-300 transition-colors">
          <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M20.317 4.3698a19.7913 19.7913 0 00-4.8851-1.5152.0741.0741 0 00-.0785.0371c-.211.3753-.4447.8648-.6083 1.2495-1.8447-.2762-3.68-.2762-5.4868 0-.1636-.3933-.4058-.8742-.6177-1.2495a.077.077 0 00-.0785-.037 19.7363 19.7363 0 00-4.8852 1.515.0699.0699 0 00-.0321.0277C.5334 9.0458-.319 13.5799.0992 18.0578a.0824.0824 0 00.0312.0561c2.0528 1.5076 4.0413 2.4228 5.9929 3.0294a.0777.0777 0 00.0842-.0276c.4616-.6304.8731-1.2952 1.226-1.9942a.076.076 0 00-.0416-.1057c-.6528-.2476-1.2743-.5495-1.8722-.8923a.077.077 0 01-.0076-.1277c.1258-.0943.2517-.1923.3718-.2914a.0743.0743 0 01.0776-.0105c3.9278 1.7933 8.18 1.7933 12.0614 0a.0739.0739 0 01.0785.0095c.1202.099.246.1981.3728.2924a.077.077 0 01-.0066.1276 12.2986 12.2986 0 01-1.873.8914.0766.0766 0 00-.0407.1067c.3604.698.7719 1.3628 1.225 1.9932a.076.076 0 00.0842.0286c1.961-.6067 3.9495-1.5219 6.0023-3.0294a.077.077 0 00.0313-.0552c.5004-5.177-.8382-9.6739-3.5485-13.6604a.061.061 0 00-.0312-.0286zM8.02 15.3312c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9555-2.4189 2.157-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.9555 2.4189-2.1569 2.4189zm7.9748 0c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9554-2.4189 2.1569-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.946 2.4189-2.1568 2.4189z" />
          </svg>
          For bug reports please use our discord.
        </Link>
        <Link href="https://patreon.com/wtrlab" className="flex items-center gap-2 px-5 py-2.5 bg-[#1e1e3a] hover:bg-[#2a2a4a] rounded-lg text-sm text-gray-300 transition-colors">
          <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M14.82 2.41C18.78 2.41 22 5.65 22 9.62C22 13.58 18.78 16.8 14.82 16.8C10.85 16.8 7.61 13.58 7.61 9.62C7.61 5.65 10.85 2.41 14.82 2.41M2 21.59H5.81V2.41H2V21.59Z" />
          </svg>
          Do you like this site? Support us.
        </Link>
      </div>

      {/* Recent Updates + Latest News + Top Spenders Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Recent Updates */}
        <div className="lg:col-span-2">
          <SectionHeader title="Recent Updates" />
          <div className="bg-[#12122a] border border-[#1e1e3a] rounded-xl p-4 max-h-[600px] overflow-y-auto">
            {recentUpdates.map((item, i) => (
              <UpdateItem key={i} {...item} />
            ))}
            <button className="w-full text-center text-sm text-violet-400 hover:text-violet-300 py-3 transition-colors">
              Load More
            </button>
          </div>
        </div>

        {/* Right sidebar */}
        <div className="space-y-6">
          {/* Latest News */}
          <div>
            <SectionHeader title="Latest News" href="/news" />
            <div className="bg-[#12122a] border border-[#1e1e3a] rounded-xl p-4 space-y-3">
              <Link href="/news/428" className="block text-sm text-gray-200 hover:text-violet-400 transition-colors">
                🎉 16th Giveaway Winners 🎉
              </Link>
              <Link href="/news/427" className="block text-sm text-gray-200 hover:text-violet-400 transition-colors">
                🎉 Our 16th Giveaway is LIVE! 🎉
              </Link>
              <Link href="/news/426" className="block text-sm text-gray-200 hover:text-violet-400 transition-colors">
                Version 1.13.3 - New Source Management (Work in Progress) & Bug Fixes!
              </Link>
            </div>
          </div>

          {/* Daily Top Spenders */}
          <div>
            <SectionHeader title="Daily Top Spenders" href="/leaderboard" />
            <div className="bg-[#12122a] border border-[#1e1e3a] rounded-xl p-4 space-y-3">
              {topSpenders.map((spender, i) => (
                <Link key={i} href={spender.href} className="flex items-center justify-between group">
                  <span className="text-sm text-gray-200 group-hover:text-violet-400 transition-colors">{spender.name}</span>
                  <span className="text-xs text-gray-500">{spender.tickets} Tickets</span>
                </Link>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Random Novels */}
      <section>
        <SectionHeader title="Random Novels" href="/random-novels" />
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-4">
          {randomNovels.map((novel, i) => (
            <NovelCard key={i} {...novel} compact />
          ))}
        </div>
      </section>
    </div>
  );
}
