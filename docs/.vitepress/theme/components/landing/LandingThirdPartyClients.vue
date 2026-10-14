<script setup lang="ts">
import { inject, type Ref } from 'vue'
import { withBase } from 'vitepress'
import type { Lang } from '../../../data/downloads'
import { createT } from '../../../data/landing-i18n'
import { THIRD_PARTY_CLIENTS as clients } from '../../../data/third-party-clients'

const lang = inject<Ref<Lang>>('landingLang')!
const t = (k: string) => createT(lang.value)(k)

const getHref = (href: string | { zh: string; en: string }) =>
  typeof href === 'string' ? href : href[lang.value]
</script>

<template>
  <section class="thirdparty" data-reveal>
    <div class="landing-container thirdparty-inner">
      <p class="section-eyebrow">{{ t('thirdparty.eyebrow') }}</p>
      <h2 class="section-title">{{ t('thirdparty.title') }}</h2>
      <p class="section-subtitle">{{ t('thirdparty.subtitle') }}</p>

      <div class="client-cards">
        <a
          v-for="c in clients"
          :key="c.href"
          class="client-card"
          :href="getHref(c.href)"
          target="_blank"
          rel="noreferrer"
        >
          <img class="client-icon" :src="withBase(`/thirdparty/${c.img}`)" :alt="c.name[lang]" />
          <div class="client-info">
            <span class="client-name">{{ c.name[lang] }}</span>
            <span class="client-desc">{{ c.desc[lang] }}</span>
            <div class="client-platforms">
              <span v-for="p in c.platforms" :key="p" class="platform-tag">{{ p }}</span>
            </div>
          </div>
          <span class="client-arrow">→</span>
        </a>
      </div>
    </div>
  </section>
</template>

<style scoped>
.thirdparty {
  padding: 72px 0;
}
.thirdparty-inner { text-align: center; }
.client-cards {
  margin-top: 34px;
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 16px;
  max-width: 1100px;
  margin-left: auto;
  margin-right: auto;
}
.client-card {
  grid-column: span 2;
  display: flex;
  align-items: flex-start;
  gap: 14px;
  min-height: 180px;
  padding: 22px 20px;
  border: 1px solid var(--vp-c-divider);
  border-radius: 14px;
  background: var(--vp-c-bg);
  text-align: left;
  transition: all 0.18s;
}
.client-card:nth-last-child(2):nth-child(3n + 1) {
  grid-column: 2 / span 2;
}
.client-card:hover {
  border-color: var(--vp-c-brand-1);
  transform: translateY(-3px);
  box-shadow: 0 14px 30px -18px rgba(0, 0, 0, 0.4);
}
.client-icon { width: 40px; height: 40px; border-radius: 8px; flex-shrink: 0; }
.client-info { flex: 1; display: flex; flex-direction: column; gap: 4px; }
.client-name {
  font-weight: 700;
  font-size: 16px;
  color: var(--vp-c-text-1);
}
.client-desc {
  font-size: 13px;
  color: var(--vp-c-text-2);
  line-height: 1.5;
}
.client-platforms {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}
.platform-tag {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--vp-c-brand-soft);
  color: var(--vp-c-brand-1);
}
.client-arrow {
  color: var(--vp-c-brand-1);
  font-weight: 700;
  font-size: 18px;
  margin-top: 4px;
  transition: transform 0.18s;
  flex-shrink: 0;
}
.client-card:hover .client-arrow { transform: translateX(4px); }
@media (max-width: 1023px) {
  .client-cards { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .client-card,
  .client-card:nth-last-child(2):nth-child(3n + 1) {
    grid-column: auto;
    min-height: 0;
  }
}
@media (max-width: 680px) {
  .client-cards { grid-template-columns: 1fr; }
  .client-card { padding: 20px 18px; }
}
</style>
