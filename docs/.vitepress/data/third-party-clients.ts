// 社区客户端的中英介绍与链接，供落地页卡片使用。
// Songloft Desktop 图标来自其仓库 build/icons/128x128.png。
export const THIRD_PARTY_CLIENTS = [
  {
    name: { zh: 'Songloft Desktop', en: 'Songloft Desktop' },
    desc: {
      zh: 'edwinhuish 独立维护的 Electron 桌面客户端，需连接 Songloft 服务端',
      en: 'Electron desktop client maintained by edwinhuish; requires a Songloft server',
    },
    platforms: ['Windows', 'macOS', 'Linux'],
    href: 'https://github.com/edwinhuish/songloft-desktop',
    img: 'songloft-desktop.png',
  },
  {
    name: { zh: '音乐方舟', en: 'MusicArk' },
    desc: { zh: 'HarmonyOS NEXT 多源媒体播放器，16 类媒体库一站聚合', en: 'Multi-source media player for HarmonyOS NEXT with 16 library types' },
    platforms: ['HarmonyOS'],
    href: 'https://musicark.pro/',
    img: 'musicark.svg',
  },
  {
    name: { zh: '箭头音乐', en: 'Amcfy Music' },
    desc: { zh: '现代化多平台音乐播放器，兼容多种音乐服务器协议', en: 'Modern multi-platform music player compatible with various music server protocols' },
    platforms: ['Android', 'iOS', 'HarmonyOS', 'Windows', 'macOS'],
    href: 'https://www.amcfy.com/',
    img: 'amcfy.png',
  },
  {
    name: { zh: '音流', en: 'Stream Music' },
    desc: { zh: '跨平台 NAS 音乐播放器，支持多种自托管音乐服务', en: 'Cross-platform NAS music player supporting multiple self-hosted music services' },
    platforms: ['Android', 'iOS', 'macOS', 'Windows'],
    href: 'https://music.aqzscn.cn/',
    img: 'stream-music.png',
  },
  {
    name: { zh: '流云音盒', en: 'XGPlayer' },
    desc: { zh: 'HarmonyOS 无损音乐播放器，聚合本地、云盘、WebDAV 与 NAS 私有曲库', en: 'Lossless music player for HarmonyOS aggregating local, cloud, WebDAV and NAS libraries' },
    platforms: ['HarmonyOS'],
    href: 'https://xgplayer.com/',
    img: 'xgplayer.png',
  },
  {
    name: { zh: '流云音乐', en: 'Cloudflow Music' },
    desc: { zh: '网盘音乐播放器，支持 WebDAV、Navidrome 等多种云盘与自托管音乐服务', en: 'Cloud music player supporting WebDAV, Navidrome and various cloud storage services' },
    platforms: ['iOS'],
    href: { zh: 'http://music.lyzo.top/', en: 'https://ly.pyzo.top/' },
    img: 'cloudflow-music.png',
  },
]
