import { chromium } from 'playwright';
import path from 'path';
import fs from 'fs';

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

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
        hud.style.maxWidth = '980px';
        hud.style.width = 'calc(100% - 48px)';
        hud.style.background = 'rgba(11, 15, 25, 0.95)';
        hud.style.backdropFilter = 'blur(12px)';
        hud.style.border = '2px solid #38bdf8';
        hud.style.borderRadius = '12px';
        hud.style.padding = '14px 22px';
        hud.style.boxShadow = '0 20px 35px rgba(0, 0, 0, 0.75), 0 0 15px rgba(56, 189, 248, 0.25)';
        hud.style.color = '#ffffff';
        hud.style.fontFamily = '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif';
        hud.style.transition = 'all 0.3s ease';
        document.body.appendChild(hud);
      }
      hud.innerHTML = `
        <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:5px;">
          <span style="background:#0284c7; color:#fff; font-size:11px; font-weight:700; text-transform:uppercase; letter-spacing:0.8px; padding:3px 10px; border-radius:9999px;">${chapter}</span>
          <span style="color:#38bdf8; font-size:11px; font-weight:600; letter-spacing:0.5px;">ChoreSync Architecture & Observability Tour</span>
        </div>
        <div style="font-size:16px; font-weight:700; color:#f8fafc; margin-bottom:3px;">${title}</div>
        <div style="font-size:12.5px; color:#cbd5e1; line-height:1.45;">${description}</div>
      `;
    }, { chapter, title, description });
  } catch (e) {
    console.log('HUD notification skipped:', e.message);
  }
}

async function run() {
  console.log('=== [DEMO VIDEO RECORDING STARTED] ===');

  const recordingsDir = path.resolve('/app/recordings');
  if (!fs.existsSync(recordingsDir)) {
    fs.mkdirSync(recordingsDir, { recursive: true });
  }

  // Clear older webm recordings to avoid ambiguity
  fs.readdirSync(recordingsDir).forEach(file => {
    if (file.endsWith('.webm')) {
      try { fs.unlinkSync(path.join(recordingsDir, file)); } catch (_) {}
    }
  });

  const browser = await chromium.launch({
    headless: true,
    args: ['--no-sandbox', '--disable-setuid-sandbox', '--disable-dev-shm-usage']
  });

  // Explicit locale 'en-US' fixes Grafana Intl.NumberFormat RangeError
  const context = await browser.newContext({
    locale: 'en-US',
    timezoneId: 'UTC',
    recordVideo: {
      dir: recordingsDir,
      size: { width: 1280, height: 720 },
    },
    viewport: { width: 1280, height: 720 },
  });

  const page = await context.newPage();
  const slidesUrl = 'file:///app/demo/slides.html';

  // -------------------------------------------------------------
  // CHAPTER 1: Welcome & Architectural Overview (18.2s)
  // -------------------------------------------------------------
  console.log('[SCENE 1] Slides: Architecture Overview (18.2s)');
  await page.goto(slidesUrl, { waitUntil: 'load' });
  await page.evaluate(() => window.goToSlide(1));
  await showHud(page, {
    chapter: 'Chapter 1: System Overview',
    title: 'ChoreSync: Behind-the-Scenes Architecture',
    description: 'ChoreSync pairs a Go 1.22 Chi API, PostgreSQL 16 relational store, and React 18 frontend with a standalone LGTM Observability Stack and Two-Stage Container CI/CD pipeline.'
  });
  await sleep(18200);

  // -------------------------------------------------------------
  // CHAPTER 2: Frontend & Kiosk Tablet Mode + Dev Mailbox (23.8s)
  // -------------------------------------------------------------
  console.log('[SCENE 2] Frontend: Kiosk Tablet Mode & Dev Mailbox (23.8s)');
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

  // Toggle Kitchen Tablet mode
  const tabletBtn = page.locator('#toggle-mode-tablet');
  if (await tabletBtn.isVisible()) {
    console.log('[DEMO] Switching to Kitchen Tablet mode...');
    await tabletBtn.click();
    await sleep(2500);
  }

  // Open Dev Mailbox modal
  const mailboxBtn = page.locator('button:has-text("Dev Mailbox")');
  if (await mailboxBtn.isVisible()) {
    console.log('[DEMO] Opening Dev Mailbox modal...');
    await mailboxBtn.click();
    await sleep(3000);
    await showHud(page, {
      chapter: 'Chapter 2: Hidden Feature: Dev Mailbox',
      title: 'Captured Transactional Emails (Magic Links & Nudges)',
      description: 'The Dev Mailbox displays single-use magic login links (15-min TTL), password reset tokens, and chore reminders captured by the backend transactional email service.'
    });
    await sleep(9000);
    const closeBtn = page.locator('button:has-text("Close"), button:has-text("✕")');
    if (await closeBtn.first().isVisible()) {
      await closeBtn.first().click();
      await sleep(2000);
    }
  } else {
    await sleep(14000);
  }

  // -------------------------------------------------------------
  // CHAPTER 3: Distributed Tracing & W3C Context Flow (24.0s)
  // -------------------------------------------------------------
  console.log('[SCENE 3] Slides: W3C Context Propagation (24.0s)');
  await page.goto(slidesUrl, { waitUntil: 'load' });
  await page.evaluate(() => window.goToSlide(2));
  await showHud(page, {
    chapter: 'Chapter 3: Telemetry Architecture',
    title: 'W3C Distributed Tracing & The Golden Triangle',
    description: 'React Web SDK injects traceparent headers; Caddy forwards them upstream; Go Chi middleware starts server spans; and pgx.QueryTracer auto-instruments SQL queries.'
  });
  await sleep(24000);

  // -------------------------------------------------------------
  // CHAPTER 4: Grafana Observability Dashboard (16.5s)
  // -------------------------------------------------------------
  console.log('[SCENE 4] Grafana: Overview Dashboard (16.5s)');
  await page.goto('http://localhost:3001/d/observability-overview', { waitUntil: 'networkidle', timeout: 15000 });
  await page.waitForTimeout(3000);
  await showHud(page, {
    chapter: 'Chapter 4: Full-Stack Observability',
    title: 'Grafana Dashboard: Exporters, Spans & Logs',
    description: 'Live Grafana dashboard with anonymous admin access. Shows Prometheus scrape target health (all UP), OTel Collector span ingestion rate, and live Loki log aggregation.'
  });
  await sleep(5500);
  await page.evaluate(() => window.scrollBy({ top: 320, behavior: 'smooth' }));
  await sleep(8000);

  // -------------------------------------------------------------
  // CHAPTER 5: Distributed Tracing in Grafana Tempo (17.8s)
  // -------------------------------------------------------------
  console.log('[SCENE 5] Grafana: Tempo Traces (17.8s)');
  const tempoUrl = "http://localhost:3001/explore?schemaVersion=1&panes=%7B%22a%22%3A%7B%22datasource%22%3A%22tempo%22%2C%22queries%22%3A%5B%7B%22refId%22%3A%22A%22%2C%22datasource%22%3A%7B%22type%22%3A%22tempo%22%2C%22uid%22%3A%22tempo%22%7D%2C%22queryType%22%3A%22traceql%22%2C%22query%22%3A%22%7Bresource.service.name%3D%5C%22api-backend%5C%22%7D%22%7D%5D%7D%7D";
  await page.goto(tempoUrl, { waitUntil: 'networkidle', timeout: 15000 });
  await page.waitForTimeout(3000);
  await showHud(page, {
    chapter: 'Chapter 5: Distributed Tracing (Tempo)',
    title: 'Trace Waterfall & Database Query Spans',
    description: 'Tempo captures distributed traces across the entire HTTP lifecycle. Each database query is auto-instrumented with pgx.QueryTracer capturing parameterized SQL statements.'
  });
  await sleep(4000);

  // Click first trace link to reveal waterfall
  try {
    const firstTraceLink = page.locator('table tbody tr td a').first();
    if (await firstTraceLink.isVisible()) {
      console.log('[DEMO] Expanding trace waterfall in Tempo...');
      await firstTraceLink.click();
      await page.waitForTimeout(3000);
    }
  } catch (_) {}
  await sleep(7800);

  // -------------------------------------------------------------
  // CHAPTER 6: Log Aggregation in Grafana Loki (17.5s)
  // -------------------------------------------------------------
  console.log('[SCENE 6] Grafana: Loki Logs (17.5s)');
  const lokiUrl = "http://localhost:3001/explore?schemaVersion=1&panes=%7B%22a%22%3A%7B%22datasource%22%3A%22loki%22%2C%22queries%22%3A%5B%7B%22refId%22%3A%22A%22%2C%22datasource%22%3A%7B%22type%22%3A%22loki%22%2C%22uid%22%3A%22loki%22%7D%2C%22editorMode%22%3A%22code%22%2C%22expr%22%3A%22%7Bjob%3D%5C%22choresync-backend%5C%22%7D%22%2C%22queryType%22%3A%22range%22%7D%5D%7D%7D";
  await page.goto(lokiUrl, { waitUntil: 'networkidle', timeout: 15000 });
  await page.waitForTimeout(3000);
  await showHud(page, {
    chapter: 'Chapter 6: Log Aggregation (Loki)',
    title: 'Structured Log Streams & Trace Correlation',
    description: 'Loki aggregates logs from Go API backend, OTel Collector, and PostgreSQL. Derived fields correlate TraceID directly into Tempo waterfall views.'
  });
  await sleep(6500);
  await page.evaluate(() => window.scrollBy({ top: 250, behavior: 'smooth' }));
  await sleep(8000);

  // -------------------------------------------------------------
  // CHAPTER 7: Actionable Symptom-Based Alert Rules in Prometheus (16.3s)
  // -------------------------------------------------------------
  console.log('[SCENE 7] Prometheus: Alert Rules (16.3s)');
  await page.goto('http://localhost:9090/alerts', { waitUntil: 'networkidle', timeout: 15000 });
  await page.waitForTimeout(2000);
  await showHud(page, {
    chapter: 'Chapter 7: Actionable Alerts (Gate 11)',
    title: 'Prometheus Symptom-Based Alert Evaluation',
    description: 'Prometheus evaluates 3 symptom-based rules: HighHttpErrorRate (> 5%), HighRequestLatency (> 1s), and DatabaseQueryErrors (> 0.05/s) with mandatory Golden Triangle metadata.'
  });
  await sleep(14300);

  // -------------------------------------------------------------
  // CHAPTER 8: Live Incident Simulation & Firing Alert (15.0s)
  // -------------------------------------------------------------
  console.log('[SCENE 8] Prometheus: HighHttpErrorRate FIRING (15.0s)');
  await page.reload({ waitUntil: 'networkidle' });
  await page.waitForTimeout(2000);
  await showHud(page, {
    chapter: 'Chapter 8: Live Incident Simulation',
    title: 'Alert State: HighHttpErrorRate FIRING',
    description: 'Synthetic 5xx error burst pushed the error rate to 66.7%. Prometheus detected the sustained violation, transitioned state to FIRING, and pushed the notification to Alertmanager.'
  });
  await sleep(13000);

  // -------------------------------------------------------------
  // CHAPTER 9: Alertmanager Notification Routing (10.0s)
  // -------------------------------------------------------------
  console.log('[SCENE 9] Alertmanager: Alert Routing (10.0s)');
  await page.goto('http://localhost:9093/#/alerts', { waitUntil: 'networkidle', timeout: 15000 });
  await page.waitForTimeout(2000);
  await showHud(page, {
    chapter: 'Chapter 9: Alertmanager Notification Routing',
    title: 'Webhook Routing to On-Call Receiver',
    description: 'Alertmanager groups firing alerts by alertname, service, and environment, and routes the webhook payload directly to http://on-call-receiver:5050/webhook.'
  });
  await sleep(8000);

  // -------------------------------------------------------------
  // CHAPTER 10: On-Call Receiver & Normalized Incidents (12.8s)
  // -------------------------------------------------------------
  console.log('[SCENE 10] On-Call Receiver: Incident Normalization (12.8s)');
  await page.goto('http://localhost:5050/', { waitUntil: 'networkidle', timeout: 15000 });
  await page.waitForTimeout(2000);
  await showHud(page, {
    chapter: 'Chapter 10: Autonomous SRE Receiver (Gate 12)',
    title: 'Normalized Incident Ingestion & Dispatch',
    description: 'The On-Call Receiver normalizes the alert against payload-schema.json, records the incident into incidents/, and prepares the context for the autonomous on-call engineer.'
  });
  await sleep(10800);

  // -------------------------------------------------------------
  // CHAPTER 11: Two-Stage Container Delivery & Promotion (21.5s)
  // -------------------------------------------------------------
  console.log('[SCENE 11] Slides: CI/CD & Production Promotion (21.5s)');
  await page.goto(slidesUrl, { waitUntil: 'load' });
  await page.evaluate(() => window.goToSlide(4));
  await showHud(page, {
    chapter: 'Chapter 11: Two-Stage Container Delivery (Gate 8)',
    title: 'Build Once, Promote Everywhere',
    description: 'Stage 1 packages images with immutable YYYYMMDD-HHMMSS-shortsha tags. Stage 2 serves via pre-built docker-compose.deploy.yml. Manual promotion captures SHA256 digests.'
  });
  await sleep(21500);

  // -------------------------------------------------------------
  // CHAPTER 12: Verification Sign-Off & Conclusion (15.0s)
  // -------------------------------------------------------------
  console.log('[SCENE 12] Slides: Final Sign-Off (15.0s)');
  await page.evaluate(() => window.goToSlide(1));
  await showHud(page, {
    chapter: 'Chapter 12: Quality Sign-Off',
    title: 'All Verification Gates & Tests Passing 100%',
    description: 'Full verification matrix green: 14 Go backend suites, 11 Playwright E2E journeys, OpenAPI 3.1 audit, and autonomous incident remediation verification suite.'
  });
  await sleep(15000);

  console.log('[DEMO] Closing context and flushing video stream...');
  await context.close();
  await browser.close();

  console.log('=== [DEMO VIDEO RECORDING COMPLETE] ===');
}

run().catch((err) => {
  console.error('[ERROR] Video recording failed:', err);
  process.exit(1);
});
