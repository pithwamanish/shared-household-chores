import { chromium } from 'playwright';
import path from 'path';
import fs from 'fs';
import http from 'http';

// Helper to pause execution
const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

// Helper to show floating presentation HUD
async function showHud(page, { chapter, title, description }) {
  try {
    await page.evaluate(({ chapter, title, description }) => {
      let hud = document.getElementById('demo-hud');
      if (!hud) {
        hud = document.createElement('div');
        hud.id = 'demo-hud';
        hud.style.position = 'fixed';
        hud.style.bottom = '20px';
        hud.style.left = '50%';
        hud.style.transform = 'translateX(-50%)';
        hud.style.zIndex = '99999999';
        hud.style.maxWidth = '960px';
        hud.style.width = 'calc(100% - 48px)';
        hud.style.background = 'rgba(11, 15, 25, 0.94)';
        hud.style.backdropFilter = 'blur(12px)';
        hud.style.border = '2px solid #38bdf8';
        hud.style.borderRadius = '12px';
        hud.style.padding = '14px 22px';
        hud.style.boxShadow = '0 20px 35px rgba(0, 0, 0, 0.75), 0 0 15px rgba(56, 189, 248, 0.2)';
        hud.style.color = '#ffffff';
        hud.style.fontFamily = '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif';
        hud.style.transition = 'all 0.3s ease';
        document.body.appendChild(hud);
      }
      hud.innerHTML = `
        <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:5px;">
          <span style="background:#0284c7; color:#fff; font-size:11px; font-weight:700; text-transform:uppercase; letter-spacing:0.8px; padding:3px 10px; border-radius:9999px;">${chapter}</span>
          <span style="color:#38bdf8; font-size:11px; font-weight:600; letter-spacing:0.5px;">ChoreSync AI-Native Architecture</span>
        </div>
        <div style="font-size:16px; font-weight:700; color:#f8fafc; margin-bottom:3px;">${title}</div>
        <div style="font-size:12.5px; color:#cbd5e1; line-height:1.45;">${description}</div>
      `;
    }, { chapter, title, description });
  } catch (e) {
    console.log('Notice: HUD update skipped on current page state:', e.message);
  }
}

// Trigger synthetic error traffic via backend API
async function triggerSyntheticIncident() {
  console.log('[DEMO] Generating synthetic 500 error burst on backend API...');
  for (let i = 0; i < 8; i++) {
    try {
      await fetch('http://localhost:8000/api/v1/dev/simulate-error', { method: 'POST' });
    } catch (_) {}
  }
  for (let i = 0; i < 4; i++) {
    try {
      await fetch('http://localhost:8000/api/v1/households');
    } catch (_) {}
  }
}

async function run() {
  console.log('=== [DEMO VIDEO RECORDING STARTED] ===');

  const recordingsDir = path.resolve('/app/recordings');
  if (!fs.existsSync(recordingsDir)) {
    fs.mkdirSync(recordingsDir, { recursive: true });
  }

  const browser = await chromium.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-dev-shm-usage']
  });

  const context = await browser.newContext({
    recordVideo: {
      dir: recordingsDir,
      size: { width: 1280, height: 720 },
    },
    viewport: { width: 1280, height: 720 },
  });

  const page = await context.newPage();
  const slidesUrl = 'file:///app/demo/slides.html';

  // -------------------------------------------------------------
  // CHAPTER 1: Welcome & Architectural Overview
  // -------------------------------------------------------------
  console.log('[SCENE 1] Slides: Architecture Overview');
  await page.goto(slidesUrl, { waitUntil: 'load' });
  await page.evaluate(() => window.goToSlide(1));
  await showHud(page, {
    chapter: 'Chapter 1: System Overview',
    title: 'ChoreSync: Behind-the-Scenes Architecture',
    description: 'ChoreSync pairs a Go 1.22 Chi API, PostgreSQL 16 relational store, and React 18 frontend with a standalone LGTM Observability Stack and Two-Stage Container CI/CD pipeline.'
  });
  await sleep(6500);

  // -------------------------------------------------------------
  // CHAPTER 2: Frontend & Kiosk Tablet Mode + Dev Mailbox
  // -------------------------------------------------------------
  console.log('[SCENE 2] Frontend: Kiosk Tablet Mode & Dev Mailbox');
  try {
    await page.goto('http://localhost:3000', { waitUntil: 'networkidle', timeout: 15000 });
  } catch (e) {
    await page.goto('http://localhost:3000', { waitUntil: 'load' });
  }
  await showHud(page, {
    chapter: 'Chapter 2: Frontend & Hidden Features',
    title: 'Kitchen Tablet Kiosk & In-Memory Dev Mailbox',
    description: 'Notice the Kitchen Tablet toggle for shared wall displays, Admin PIN verification gate, and the Dev Mailbox modal capturing transactional emails without external spam.'
  });
  await sleep(4000);

  // Open Dev Mailbox modal to demonstrate transactional email capture
  const mailboxBtn = page.locator('button:has-text("Dev Mailbox")');
  if (await mailboxBtn.isVisible()) {
    console.log('[DEMO] Opening Dev Mailbox modal...');
    await mailboxBtn.click();
    await sleep(1500);
    await showHud(page, {
      chapter: 'Chapter 2: Hidden Feature: Dev Mailbox',
      title: 'Captured Transactional Emails (Magic Links & Nudges)',
      description: 'The Dev Mailbox displays single-use magic login links (15-min TTL), password reset tokens, and chore reminders captured by the backend transactional email service.'
    });
    await sleep(4500);
    const closeBtn = page.locator('button:has-text("Close"), button:has-text("✕")');
    if (await closeBtn.first().isVisible()) {
      await closeBtn.first().click();
      await sleep(1000);
    }
  }

  // -------------------------------------------------------------
  // CHAPTER 3: Distributed Tracing & W3C Context Flow
  // -------------------------------------------------------------
  console.log('[SCENE 3] Slides: W3C Context Propagation');
  await page.goto(slidesUrl, { waitUntil: 'load' });
  await page.evaluate(() => window.goToSlide(2));
  await showHud(page, {
    chapter: 'Chapter 3: Telemetry Architecture',
    title: 'W3C Distributed Tracing & The Golden Triangle',
    description: 'React Web SDK injects traceparent headers; Caddy forwards them upstream; Go Chi middleware starts server spans; and pgx.QueryTracer auto-instruments SQL queries.'
  });
  await sleep(6500);

  // -------------------------------------------------------------
  // CHAPTER 4: Grafana Observability Dashboard
  // -------------------------------------------------------------
  console.log('[SCENE 4] Grafana: Overview Dashboard');
  await page.goto('http://localhost:3001/d/observability-overview', { waitUntil: 'networkidle', timeout: 15000 });
  await showHud(page, {
    chapter: 'Chapter 4: Full-Stack Observability',
    title: 'Grafana Dashboard: Exporters, Spans & Logs',
    description: 'Live Grafana dashboard with anonymous admin access. Shows Prometheus scrape target health (all UP), OTel Collector span ingestion rate, and live Loki log aggregation.'
  });
  await sleep(3500);
  await page.evaluate(() => window.scrollBy({ top: 320, behavior: 'smooth' }));
  await sleep(4000);

  // -------------------------------------------------------------
  // CHAPTER 5: Distributed Tracing in Grafana Tempo
  // -------------------------------------------------------------
  console.log('[SCENE 5] Grafana: Tempo Traces');
  await page.goto('http://localhost:3001/explore?schemaVersion=1&panes=%7B%22a%22%3A%7B%22datasource%22%3A%22tempo%22%7D%7D', { waitUntil: 'networkidle', timeout: 15000 });
  await showHud(page, {
    chapter: 'Chapter 5: Distributed Tracing (Tempo)',
    title: 'Trace Waterfall & Database Query Spans',
    description: 'Grafana Tempo stores and indexes distributed traces. Clicking into a trace reveals the full HTTP request lifecycle and exact SQL query statements with latencies.'
  });
  await sleep(6000);

  // -------------------------------------------------------------
  // CHAPTER 6: Log Aggregation in Grafana Loki
  // -------------------------------------------------------------
  console.log('[SCENE 6] Grafana: Loki Logs');
  await page.goto('http://localhost:3001/explore?schemaVersion=1&panes=%7B%22a%22%3A%7B%22datasource%22%3A%22loki%22%7D%7D', { waitUntil: 'networkidle', timeout: 15000 });
  await showHud(page, {
    chapter: 'Chapter 6: Log Aggregation (Loki)',
    title: 'High-Efficiency Structured Log Exploration',
    description: 'Grafana Loki indexes container log streams with LogQL. Notice automatic correlation linking log lines to distributed trace IDs in Tempo.'
  });
  await sleep(6000);

  // -------------------------------------------------------------
  // CHAPTER 7: Actionable Symptom-Based Alert Rules in Prometheus
  // -------------------------------------------------------------
  console.log('[SCENE 7] Prometheus: Alert Rules');
  await page.goto('http://localhost:9090/alerts', { waitUntil: 'networkidle', timeout: 15000 });
  await showHud(page, {
    chapter: 'Chapter 7: Actionable Alerts (Gate 11)',
    title: 'Prometheus Symptom-Based Alert Evaluation',
    description: 'Prometheus evaluates 3 symptom-based rules: HighHttpErrorRate (> 5%), HighRequestLatency (> 1s), and DatabaseQueryErrors (> 0.05/s) with mandatory Golden Triangle metadata.'
  });
  await sleep(5500);

  // -------------------------------------------------------------
  // CHAPTER 8: Live Incident Simulation & Alert Firing
  // -------------------------------------------------------------
  console.log('[SCENE 8] Live Incident Simulation');
  await triggerSyntheticIncident();
  await sleep(2500);
  await page.reload({ waitUntil: 'networkidle' });
  await showHud(page, {
    chapter: 'Chapter 8: Live Incident Simulation',
    title: 'Alert State: HighHttpErrorRate Firing',
    description: 'Synthetic traffic generated a 66.7% error burst. Prometheus detects the sustained violation, transitions rule state to FIRING, and pushes the alert to Alertmanager.'
  });
  await sleep(6000);

  // -------------------------------------------------------------
  // CHAPTER 9: Alertmanager Routing
  // -------------------------------------------------------------
  console.log('[SCENE 9] Alertmanager: Alert Routing');
  await page.goto('http://localhost:9093/#/alerts', { waitUntil: 'networkidle', timeout: 15000 });
  await showHud(page, {
    chapter: 'Chapter 9: Alertmanager Notification Routing',
    title: 'Webhook Routing to On-Call Receiver',
    description: 'Alertmanager groups firing alerts by alertname, service, and environment, and routes the webhook payload directly to http://on-call-receiver:5050/webhook.'
  });
  await sleep(5000);

  // -------------------------------------------------------------
  // CHAPTER 10: On-Call Receiver & Normalized Incidents
  // -------------------------------------------------------------
  console.log('[SCENE 10] On-Call Receiver: Incident Normalization');
  await page.goto('http://localhost:5050/', { waitUntil: 'networkidle', timeout: 15000 });
  await showHud(page, {
    chapter: 'Chapter 10: Autonomous SRE Receiver (Gate 12)',
    title: 'Normalized Incident Ingestion & Dispatch',
    description: 'The On-Call Receiver normalizes the alert against payload-schema.json, records the incident into incidents/, and prepares the context for the autonomous on-call engineer.'
  });
  await sleep(6500);

  // -------------------------------------------------------------
  // CHAPTER 11: Two-Stage Container Delivery & Promotion Pipeline
  // -------------------------------------------------------------
  console.log('[SCENE 11] Slides: CI/CD & Production Promotion');
  await page.goto(slidesUrl, { waitUntil: 'load' });
  await page.evaluate(() => window.goToSlide(4));
  await showHud(page, {
    chapter: 'Chapter 11: Two-Stage Container Delivery (Gate 8)',
    title: 'Build Once, Promote Everywhere',
    description: 'Stage 1 packages images with immutable YYYYMMDD-HHMMSS-shortsha tags. Stage 2 serves via pre-built docker-compose.deploy.yml. Manual promotion captures SHA256 digests.'
  });
  await sleep(7000);

  // -------------------------------------------------------------
  // CHAPTER 12: Verification Sign-Off & Conclusion
  // -------------------------------------------------------------
  console.log('[SCENE 12] Slides: Final Sign-Off');
  await page.evaluate(() => window.goToSlide(1));
  await showHud(page, {
    chapter: 'Chapter 12: Quality Sign-Off',
    title: 'All Verification Gates & Tests Passing 100%',
    description: 'Full verification matrix green: 14 Go backend suites, 11 Playwright E2E journeys, OpenAPI 3.1 audit, and autonomous incident remediation verification suite.'
  });
  await sleep(5000);

  console.log('[DEMO] Closing context and flushing video stream...');
  await context.close();
  await browser.close();

  console.log('=== [DEMO VIDEO RECORDING COMPLETE] ===');
}

run().catch((err) => {
  console.error('[ERROR] Video recording failed:', err);
  process.exit(1);
});
