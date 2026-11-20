/**
 * BMKG / Windy-style HTML5 Canvas Weather Overlay Engine for Fleet Map (AQMS).
 *
 * Renders two independently toggleable 60 FPS animated layers synced to a Leaflet map:
 *  1. Wind Streamlines: Smooth IDW-interpolated vector field from real AQMS tenant
 *     `wind_angle` (0-360°) & `wind_speed`, with color/length/speed scaling and
 *     directional chevron tips.
 *  2. Wind-Slanted Rain Zones: Translucent blue radar cells at raining AQMS sites
 *     (`rain_fall > 0.05 mm`) with animated teardrop streaks whose slant & drift
 *     follow each station's exact `wind_angle` & `wind_speed`.
 */

export function normalizeWindSpeedMs(speed, unit = 'm/s') {
  const num = Number(speed);
  if (!Number.isFinite(num) || num < 0) return 0;
  const u = String(unit || 'm/s').toLowerCase();
  if (u.includes('km/h') || u.includes('kmh') || u.includes('kph')) return num / 3.6;
  if (u.includes('knot') || u === 'kt') return num * 0.514444;
  if (u.includes('mph')) return num * 0.44704;
  return num;
}

/**
 * Converts compass/meteorological angle (0°=North/Up, 90°=East/Right,
 * 180°=South/Down, 270°=West/Left) into screen-space unit/velocity vector (ux, uy).
 */
export function windAngleToCanvasVelocity(angleDeg, magnitude = 1) {
  const num = Number(angleDeg);
  if (!Number.isFinite(num)) {
    return { ux: 0, uy: 0, valid: false };
  }
  const norm = ((num % 360) + 360) % 360;
  const rad = (norm * Math.PI) / 180;
  return {
    ux: Math.sin(rad) * magnitude,
    uy: -Math.cos(rad) * magnitude,
    valid: true,
    deg: norm,
  };
}

export function resolveWindVisualStyle(speedMs, isDark = false) {
  const ms = Number.isFinite(Number(speedMs)) ? Math.max(0, Number(speedMs)) : 0;
  if (ms > 10) {
    return {
      tier: 'VERY_STRONG',
      rgb: isDark ? [251, 113, 133] : [225, 29, 72],
      hex: isDark ? '#fb7185' : '#e11d48',
      stepPx: 1.75,
      trailLen: 46,
      lineWidth: 2.25,
    };
  }
  if (ms >= 5) {
    return {
      tier: 'STRONG',
      rgb: isDark ? [251, 191, 36] : [217, 119, 6],
      hex: isDark ? '#fbbf24' : '#d97706',
      stepPx: 1.08 + ((ms - 5) / 5) * 0.56,
      trailLen: 40,
      lineWidth: 2.0,
    };
  }
  if (ms >= 2) {
    return {
      tier: 'MODERATE',
      rgb: isDark ? [52, 211, 153] : [5, 150, 105],
      hex: isDark ? '#34d399' : '#059669',
      stepPx: 0.62 + ((ms - 2) / 3) * 0.42,
      trailLen: 34,
      lineWidth: 1.75,
    };
  }
  return {
    tier: 'LIGHT',
    rgb: isDark ? [56, 189, 248] : [2, 132, 199],
    hex: isDark ? '#38bdf8' : '#0284c7',
    stepPx: 0.34 + (ms / 2) * 0.24,
    trailLen: 28,
    lineWidth: 1.5,
  };
}

export function classifyRainIntensity(rainFallMm) {
  const mm = Number.isFinite(Number(rainFallMm)) ? Math.max(0, Number(rainFallMm)) : 0;
  if (mm > 10) {
    return {
      level: 'HEAVY',
      label: 'Lebat',
      baseRadius: 68,
      dropCount: 34,
      streakLen: 15,
      dropSpeed: 0.85,
      coreAlpha: 0.40,
      midAlpha: 0.22,
      dropWidth: 2.1,
    };
  }
  if (mm >= 2) {
    return {
      level: 'MODERATE',
      label: 'Sedang',
      baseRadius: 54,
      dropCount: 22,
      streakLen: 12,
      dropSpeed: 0.62,
      coreAlpha: 0.32,
      midAlpha: 0.16,
      dropWidth: 1.8,
    };
  }
  return {
    level: 'LIGHT',
    label: 'Ringan',
    baseRadius: 42,
    dropCount: 14,
    streakLen: 8.5,
    dropSpeed: 0.42,
    coreAlpha: 0.24,
    midAlpha: 0.11,
    dropWidth: 1.5,
  };
}

/**
 * Pure Inverse-Distance-Weighted (IDW) wind field evaluation at pixel (px, py)
 * from an array of station points `{ x, y, u, v, speedMs }`.
 */
export function sampleWindFieldAtPixel(px, py, windStations, influenceRadius = 240) {
  if (!Array.isArray(windStations) || windStations.length === 0) {
    return { u: 0, v: 0, speedMs: 0, influence: 0 };
  }
  let wSum = 0;
  let uSum = 0;
  let vSum = 0;
  let speedSum = 0;
  let minDistSq = Infinity;
  const eps = 240; // Smooth kernel floor (px^2)

  for (let i = 0; i < windStations.length; i += 1) {
    const st = windStations[i];
    const dx = px - st.x;
    const dy = py - st.y;
    const d2 = dx * dx + dy * dy;
    if (d2 < minDistSq) minDistSq = d2;
    // Distance power ~2.6 gives crisp local fidelity while curving smoothly between cities
    const w = 1 / Math.pow(d2 + eps, 1.3);
    wSum += w;
    uSum += st.u * w;
    vSum += st.v * w;
    speedSum += st.speedMs * w;
  }

  if (wSum <= 0) {
    return { u: 0, v: 0, speedMs: 0, influence: 0 };
  }

  const u = uSum / wSum;
  const v = vSum / wSum;
  const mag = Math.hypot(u, v);
  const speedMs = speedSum / wSum;

  const minDist = Math.sqrt(minDistSq);
  const outerMax = influenceRadius * 1.75;
  let influence = 1;
  if (minDist > influenceRadius * 0.55) {
    const t = Math.min(1, (minDist - influenceRadius * 0.55) / (outerMax - influenceRadius * 0.55));
    // Smoothstep fade to 0 outside reach
    influence = Math.max(0, 1 - t * t * (3 - 2 * t));
  }

  return {
    u: mag > 1e-4 ? u / mag : 0,
    v: mag > 1e-4 ? v / mag : 0,
    speedMs,
    influence,
  };
}

export class FleetWeatherCanvasOverlay {
  constructor(map, options = {}) {
    this.map = map;
    this.L = options.L || null;
    this.showWind = Boolean(options.showWind);
    this.showRain = Boolean(options.showRain);
    this.isDark = Boolean(options.isDark);
    this.sites = [];

    this.canvas = null;
    this.ctx = null;
    this.pane = null;
    this.width = 0;
    this.height = 0;
    this.dpr = 1;

    // Precomputed vector field grid
    this.gridStep = 20;
    this.gridCols = 0;
    this.gridRows = 0;
    this.gridU = null;
    this.gridV = null;
    this.gridSpeed = null;
    this.gridInfluence = null;

    // Projected AQMS station caches
    this.windStations = [];
    this.rainStations = [];
    this.particles = [];

    // Animation loop state
    this.rafId = null;
    this.lastFrameTime = 0;
    this.pulsePhase = 0;
    this.isPausedByZoom = false;
    this.isPageHidden = false;
    this.destroyed = false;

    this._onMoveEnd = () => this.rebuildField();
    this._onZoomStart = () => {
      this.isPausedByZoom = true;
      this.clearCanvas();
    };
    this._onZoomEnd = () => {
      this.isPausedByZoom = false;
      this.rebuildField();
    };
    this._onResize = () => this.rebuildField();
    this._onVisibility = () => {
      this.isPageHidden = typeof document !== 'undefined' && document.hidden;
      if (!this.isPageHidden) {
        this.lastFrameTime = 0;
        this.ensureLoop();
      } else {
        this.stopLoop();
      }
    };

    this.initCanvas();
  }

  initCanvas() {
    if (!this.map || typeof document === 'undefined') return;
    let pane = this.map.getPane('fleetWeatherPane');
    if (!pane) {
      pane = this.map.createPane('fleetWeatherPane');
      pane.style.zIndex = '460';
      pane.style.pointerEvents = 'none';
    }
    this.pane = pane;

    const canvas = document.createElement('canvas');
    canvas.className = 'leaflet-layer fleet-weather-canvas-layer';
    canvas.style.position = 'absolute';
    canvas.style.top = '0';
    canvas.style.left = '0';
    canvas.style.pointerEvents = 'none';
    pane.appendChild(canvas);

    this.canvas = canvas;
    this.ctx = canvas.getContext('2d');

    this.map.on('moveend', this._onMoveEnd);
    this.map.on('zoomstart', this._onZoomStart);
    this.map.on('zoomend', this._onZoomEnd);
    this.map.on('resize', this._onResize);
    document.addEventListener('visibilitychange', this._onVisibility);

    this.rebuildField();
  }

  updateConfig({ sites, showWind, showRain, isDark }) {
    if (Array.isArray(sites)) this.sites = sites;
    if (typeof showWind === 'boolean') this.showWind = showWind;
    if (typeof showRain === 'boolean') this.showRain = showRain;
    if (typeof isDark === 'boolean') this.isDark = isDark;

    if (!this.showWind && !this.showRain) {
      this.stopLoop();
      this.clearCanvas();
      return;
    }
    this.rebuildField();
    this.ensureLoop();
  }

  syncCanvasBounds() {
    if (!this.map || !this.canvas) return false;
    const size = this.map.getSize();
    if (!size || size.x <= 0 || size.y <= 0) return false;

    const dpr = Math.min(window.devicePixelRatio || 1, 2);
    this.dpr = dpr;
    this.width = size.x;
    this.height = size.y;

    this.canvas.width = Math.round(size.x * dpr);
    this.canvas.height = Math.round(size.y * dpr);
    this.canvas.style.width = `${size.x}px`;
    this.canvas.style.height = `${size.y}px`;

    const topLeft = this.map.containerPointToLayerPoint([0, 0]);
    if (this.L && this.L.DomUtil && this.L.DomUtil.setPosition) {
      this.L.DomUtil.setPosition(this.canvas, topLeft);
    } else {
      this.canvas.style.transform = `translate3d(${topLeft.x}px, ${topLeft.y}px, 0)`;
    }

    if (this.ctx) {
      this.ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    }
    return true;
  }

  rebuildField() {
    if (this.destroyed || !this.map || !this.canvas) return;
    if (!this.showWind && !this.showRain) {
      this.stopLoop();
      this.clearCanvas();
      return;
    }
    if (!this.syncCanvasBounds()) return;

    const zoom = this.map.getZoom();
    const pad = 260;
    const windStations = [];
    const rainStations = [];

    for (let i = 0; i < this.sites.length; i += 1) {
      const s = this.sites[i];
      if (!s || String(s.monitoring_domain || '').toUpperCase() !== 'AQMS') continue;
      if (!s.weather || !s.weather.available) continue;
      const lat = Number(s.latitude);
      const lng = Number(s.longitude);
      if (!Number.isFinite(lat) || !Number.isFinite(lng)) continue;

      const pt = this.map.latLngToContainerPoint([lat, lng]);
      if (!pt || !Number.isFinite(pt.x) || !Number.isFinite(pt.y)) continue;

      const w = s.weather;
      const inBroadView =
        pt.x >= -pad && pt.x <= this.width + pad && pt.y >= -pad && pt.y <= this.height + pad;

      if (w.wind_angle != null || w.wind_speed != null) {
        const speedMs = normalizeWindSpeedMs(w.wind_speed != null ? w.wind_speed : 1.5, w.wind_speed_unit);
        const vel = windAngleToCanvasVelocity(w.wind_angle != null ? w.wind_angle : 90, 1);
        if (vel.valid) {
          windStations.push({
            siteId: s.site_id,
            x: pt.x,
            y: pt.y,
            u: vel.ux,
            v: vel.uy,
            angleDeg: vel.deg,
            speedMs,
            inView: pt.x >= -40 && pt.x <= this.width + 40 && pt.y >= -40 && pt.y <= this.height + 40,
          });
        }
      }

      if (this.showRain && w.is_raining && inBroadView) {
        const intensity = classifyRainIntensity(w.rain_fall);
        const zoomScale = Math.max(0.78, Math.min(1.45, 0.72 + zoom * 0.065));
        const radius = Math.round(intensity.baseRadius * zoomScale);

        let rainDirU = 0;
        let rainDirV = 1; // fallback straight downward
        let hasWindAngle = false;
        let windSpeedMs = 0;
        if (w.wind_angle != null) {
          const vel = windAngleToCanvasVelocity(w.wind_angle, 1);
          if (vel.valid) {
            rainDirU = vel.ux;
            rainDirV = vel.uy;
            hasWindAngle = true;
          }
        }
        if (w.wind_speed != null) {
          windSpeedMs = normalizeWindSpeedMs(w.wind_speed, w.wind_speed_unit);
        }

        rainStations.push({
          siteId: s.site_id,
          x: pt.x,
          y: pt.y,
          rainFall: Number(w.rain_fall) || 0,
          intensity,
          radius,
          dirU: rainDirU,
          dirV: rainDirV,
          hasWindAngle,
          windSpeedMs,
          drops: this.createRainDrops(intensity.dropCount, radius),
        });
      }
    }

    this.windStations = windStations;

    // Build precomputed IDW wind grid for silky-smooth BMKG streamlines
    const influenceRadius = Math.max(155, Math.min(340, 115 + zoom * 24));
    this.gridStep = this.width * this.height > 1200000 ? 22 : 18;
    this.gridCols = Math.ceil(this.width / this.gridStep) + 1;
    this.gridRows = Math.ceil(this.height / this.gridStep) + 1;
    const totalCells = this.gridCols * this.gridRows;

    if (this.showWind && windStations.length > 0) {
      this.gridU = new Float32Array(totalCells);
      this.gridV = new Float32Array(totalCells);
      this.gridSpeed = new Float32Array(totalCells);
      this.gridInfluence = new Float32Array(totalCells);

      for (let r = 0; r < this.gridRows; r += 1) {
        const py = r * this.gridStep;
        const rowOffset = r * this.gridCols;
        for (let c = 0; c < this.gridCols; c += 1) {
          const px = c * this.gridStep;
          const sample = sampleWindFieldAtPixel(px, py, windStations, influenceRadius);
          const idx = rowOffset + c;
          this.gridU[idx] = sample.u;
          this.gridV[idx] = sample.v;
          this.gridSpeed[idx] = sample.speedMs;
          this.gridInfluence[idx] = sample.influence;
        }
      }

      // If any raining station lacks its own wind_angle, borrow local interpolated wind direction!
      for (let i = 0; i < rainStations.length; i += 1) {
        const rs = rainStations[i];
        if (!rs.hasWindAngle) {
          const fld = this.lookupGrid(rs.x, rs.y);
          if (fld.influence > 0.15 && Math.hypot(fld.u, fld.v) > 0.2) {
            rs.dirU = fld.u;
            rs.dirV = fld.v;
            rs.windSpeedMs = fld.speedMs;
          }
        }
      }

      this.initParticles();
    } else {
      this.gridU = null;
      this.gridV = null;
      this.gridSpeed = null;
      this.gridInfluence = null;
      this.particles = [];
    }

    this.rainStations = rainStations;

    if ((this.showWind && this.windStations.length > 0) || (this.showRain && this.rainStations.length > 0)) {
      this.ensureLoop();
    } else {
      this.stopLoop();
      this.clearCanvas();
    }
  }

  createRainDrops(count, radius) {
    const drops = new Array(count);
    for (let i = 0; i < count; i += 1) {
      // Uniform distribution across the circular rain cell in local coordinates (-1..1)
      const angle = Math.random() * Math.PI * 2;
      const dist = Math.sqrt(Math.random()) * 0.92;
      drops[i] = {
        // Local offset perpendicular & parallel to wind direction
        perp: (Math.random() * 2 - 1) * radius * 0.86,
        prog: (Math.random() * 2 - 1) * radius * 0.9,
        speedFactor: 0.78 + Math.random() * 0.48,
        lenFactor: 0.75 + Math.random() * 0.5,
        splashPhase: Math.random() * Math.PI * 2,
        seedX: Math.cos(angle) * dist * radius * 0.7,
        seedY: Math.sin(angle) * dist * radius * 0.7,
      };
    }
    return drops;
  }

  lookupGrid(px, py) {
    if (!this.gridU || this.gridCols < 2 || this.gridRows < 2) {
      return { u: 0, v: 0, speedMs: 0, influence: 0 };
    }
    const fx = px / this.gridStep;
    const fy = py / this.gridStep;
    const c0 = Math.floor(fx);
    const r0 = Math.floor(fy);
    if (c0 < 0 || r0 < 0 || c0 >= this.gridCols - 1 || r0 >= this.gridRows - 1) {
      return { u: 0, v: 0, speedMs: 0, influence: 0 };
    }
    const tx = fx - c0;
    const ty = fy - r0;
    const i00 = r0 * this.gridCols + c0;
    const i10 = i00 + 1;
    const i01 = i00 + this.gridCols;
    const i11 = i01 + 1;

    const w00 = (1 - tx) * (1 - ty);
    const w10 = tx * (1 - ty);
    const w01 = (1 - tx) * ty;
    const w11 = tx * ty;

    const u = this.gridU[i00] * w00 + this.gridU[i10] * w10 + this.gridU[i01] * w01 + this.gridU[i11] * w11;
    const v = this.gridV[i00] * w00 + this.gridV[i10] * w10 + this.gridV[i01] * w01 + this.gridV[i11] * w11;
    const speedMs =
      this.gridSpeed[i00] * w00 +
      this.gridSpeed[i10] * w10 +
      this.gridSpeed[i01] * w01 +
      this.gridSpeed[i11] * w11;
    const influence =
      this.gridInfluence[i00] * w00 +
      this.gridInfluence[i10] * w10 +
      this.gridInfluence[i01] * w01 +
      this.gridInfluence[i11] * w11;

    const mag = Math.hypot(u, v);
    return {
      u: mag > 1e-4 ? u / mag : 0,
      v: mag > 1e-4 ? v / mag : 0,
      speedMs,
      influence,
    };
  }

  spawnParticle(randomizeAge = false) {
    const visibleStations = this.windStations.filter(s => s.inView);
    const pool = visibleStations.length > 0 ? visibleStations : this.windStations;
    let x = Math.random() * this.width;
    let y = Math.random() * this.height;

    // 78% spawn organically around AQMS stations (guaranteeing dense regional BMKG flows),
    // 22% spawn across full map to bridge wide inter-island wind currents
    if (pool.length > 0 && Math.random() < 0.78) {
      const anchor = pool[Math.floor(Math.random() * pool.length)];
      const spread = 12 + Math.pow(Math.random(), 0.72) * 185;
      const theta = Math.random() * Math.PI * 2;
      x = anchor.x + Math.cos(theta) * spread;
      y = anchor.y + Math.sin(theta) * spread;
    }

    const maxAge = 140 + Math.floor(Math.random() * 110);
    const age = randomizeAge ? Math.floor(Math.random() * maxAge) : 0;
    return {
      x,
      y,
      age,
      maxAge,
      trail: [{ x, y }],
    };
  }

  initParticles() {
    const areaFactor = (this.width * this.height) / (1280 * 720);
    const count = Math.max(220, Math.min(540, Math.round(380 * Math.sqrt(Math.max(0.4, areaFactor)))));
    this.particles = new Array(count);
    for (let i = 0; i < count; i += 1) {
      this.particles[i] = this.spawnParticle(true);
    }
  }

  clearCanvas() {
    if (!this.ctx || !this.canvas) return;
    this.ctx.clearRect(0, 0, this.width, this.height);
  }

  ensureLoop() {
    if (this.destroyed || this.isPageHidden || this.rafId != null) return;
    if (!this.showWind && !this.showRain) return;
    this.lastFrameTime = 0;
    const tick = (now) => {
      if (this.destroyed) {
        this.rafId = null;
        return;
      }
      this.rafId = window.requestAnimationFrame(tick);
      if (this.isPausedByZoom || this.isPageHidden) return;
      if (!this.lastFrameTime) this.lastFrameTime = now;
      const dtMs = Math.min(48, Math.max(8, now - this.lastFrameTime));
      this.lastFrameTime = now;
      const dtFactor = dtMs / 16.667; // normalized to 60fps
      this.renderFrame(dtFactor);
    };
    this.rafId = window.requestAnimationFrame(tick);
  }

  stopLoop() {
    if (this.rafId != null && typeof window !== 'undefined') {
      window.cancelAnimationFrame(this.rafId);
    }
    this.rafId = null;
  }

  renderFrame(dtFactor) {
    const ctx = this.ctx;
    if (!ctx || this.width <= 0 || this.height <= 0) return;

    ctx.clearRect(0, 0, this.width, this.height);
    this.pulsePhase += 0.014 * dtFactor;

    // 1. Render Rain Radar Auras underneath Wind Streamlines
    if (this.showRain && this.rainStations.length > 0) {
      this.renderRainAuras(ctx);
    }

    // 2. Render BMKG Animated Wind Streamlines
    if (this.showWind && this.particles.length > 0 && this.gridU) {
      this.renderWindStreamlines(ctx, dtFactor);
    }

    // 3. Render Wind-Slanted Animated Raindrops crisp on top of radar aura
    if (this.showRain && this.rainStations.length > 0) {
      this.renderRainDrops(ctx, dtFactor);
    }
  }

  renderRainAuras(ctx) {
    const pulse = 0.5 + 0.5 * Math.sin(this.pulsePhase);
    for (let i = 0; i < this.rainStations.length; i += 1) {
      const rs = this.rainStations[i];
      const { x, y, radius, intensity } = rs;

      // Soft translucent cobalt/sky radar mist
      const grad = ctx.createRadialGradient(x, y, 2, x, y, radius);
      if (this.isDark) {
        grad.addColorStop(0, `rgba(59, 130, 246, ${intensity.coreAlpha + 0.1})`);
        grad.addColorStop(0.52, `rgba(37, 99, 235, ${intensity.midAlpha + 0.06})`);
        grad.addColorStop(1, 'rgba(37, 99, 235, 0)');
      } else {
        grad.addColorStop(0, `rgba(37, 99, 235, ${intensity.coreAlpha})`);
        grad.addColorStop(0.55, `rgba(56, 189, 248, ${intensity.midAlpha})`);
        grad.addColorStop(1, 'rgba(56, 189, 248, 0)');
      }

      ctx.beginPath();
      ctx.arc(x, y, radius, 0, Math.PI * 2);
      ctx.fillStyle = grad;
      ctx.fill();

      // Concentric subtle pulsing radar ring for Moderate / Heavy rain
      if (intensity.level !== 'LIGHT') {
        const ringR = radius * (0.58 + pulse * 0.32);
        const ringAlpha = (1 - pulse) * (intensity.level === 'HEAVY' ? 0.52 : 0.34);
        ctx.beginPath();
        ctx.arc(x, y, ringR, 0, Math.PI * 2);
        ctx.strokeStyle = this.isDark
          ? `rgba(96, 165, 250, ${ringAlpha})`
          : `rgba(29, 78, 216, ${ringAlpha})`;
        ctx.lineWidth = intensity.level === 'HEAVY' ? 1.8 : 1.3;
        ctx.stroke();
      }
    }
  }

  renderWindStreamlines(ctx, dtFactor) {
    ctx.lineCap = 'round';
    ctx.lineJoin = 'round';

    for (let i = 0; i < this.particles.length; i += 1) {
      const p = this.particles[i];
      const fld = this.lookupGrid(p.x, p.y);

      if (
        p.age >= p.maxAge ||
        fld.influence < 0.08 ||
        p.x < -25 ||
        p.x > this.width + 25 ||
        p.y < -25 ||
        p.y > this.height + 25
      ) {
        this.particles[i] = this.spawnParticle(false);
        continue;
      }

      const style = resolveWindVisualStyle(fld.speedMs, this.isDark);
      const step = style.stepPx * dtFactor;
      const nx = p.x + fld.u * step;
      const ny = p.y + fld.v * step;

      p.x = nx;
      p.y = ny;
      p.age += 1;
      p.trail.push({ x: nx, y: ny });
      if (p.trail.length > style.trailLen) {
        p.trail.shift();
      }

      const len = p.trail.length;
      if (len < 3) continue;

      // Smooth fade-in and fade-out over particle life
      const lifeRatio = p.age / p.maxAge;
      const lifeAlpha =
        lifeRatio < 0.16
          ? lifeRatio / 0.16
          : lifeRatio > 0.72
          ? Math.max(0, (1 - lifeRatio) / 0.28)
          : 1;
      const baseAlpha = Math.min(0.92, lifeAlpha * (0.38 + 0.62 * fld.influence));
      if (baseAlpha < 0.04) continue;

      const [r, g, b] = style.rgb;
      const tail = p.trail[0];
      const head = p.trail[len - 1];

      // Tapered glowing gradient along the curved streamline trail
      const grad = ctx.createLinearGradient(tail.x, tail.y, head.x, head.y);
      grad.addColorStop(0, `rgba(${r}, ${g}, ${b}, 0)`);
      grad.addColorStop(0.45, `rgba(${r}, ${g}, ${b}, ${(baseAlpha * 0.48).toFixed(3)})`);
      grad.addColorStop(1, `rgba(${r}, ${g}, ${b}, ${baseAlpha.toFixed(3)})`);

      ctx.beginPath();
      ctx.moveTo(tail.x, tail.y);
      for (let k = 1; k < len; k += 1) {
        ctx.lineTo(p.trail[k].x, p.trail[k].y);
      }
      ctx.strokeStyle = grad;
      ctx.lineWidth = style.lineWidth;
      ctx.stroke();

      // Subtle directional chevron arrowhead at head so flow direction is unmistakable (BMKG style)
      if (baseAlpha > 0.28 && len >= 5) {
        const prev = p.trail[Math.max(0, len - 4)];
        const dx = head.x - prev.x;
        const dy = head.y - prev.y;
        const dMag = Math.hypot(dx, dy);
        if (dMag > 0.25) {
          const ux = dx / dMag;
          const uy = dy / dMag;
          const arrowLen = style.lineWidth * 2.55;
          const arrowHalfW = style.lineWidth * 1.45;
          const bx = head.x - ux * arrowLen;
          const by = head.y - uy * arrowLen;
          const pxOff = -uy * arrowHalfW;
          const pyOff = ux * arrowHalfW;

          ctx.beginPath();
          ctx.moveTo(bx + pxOff, by + pyOff);
          ctx.lineTo(head.x + ux * 1.2, head.y + uy * 1.2);
          ctx.lineTo(bx - pxOff, by - pyOff);
          ctx.strokeStyle = `rgba(${r}, ${g}, ${b}, ${Math.min(0.96, baseAlpha * 1.05).toFixed(3)})`;
          ctx.lineWidth = Math.max(1.2, style.lineWidth * 0.82);
          ctx.stroke();
        }
      }
    }
  }

  renderRainDrops(ctx, dtFactor) {
    ctx.lineCap = 'round';
    for (let i = 0; i < this.rainStations.length; i += 1) {
      const rs = this.rainStations[i];
      const { x, y, radius, intensity, dirU, dirV, windSpeedMs, drops } = rs;
      // Perpendicular unit vector to wind direction for spreading raindrops cleanly across the cell
      const perpU = -dirV;
      const perpV = dirU;
      const speedBoost = 1 + Math.min(0.65, windSpeedMs * 0.08);
      const streakBoost = 1 + Math.min(0.5, windSpeedMs * 0.05);

      for (let j = 0; j < drops.length; j += 1) {
        const d = drops[j];
        d.prog += intensity.dropSpeed * d.speedFactor * speedBoost * dtFactor;

        // Compute max travel extent inside circular rain cell for this perpendicular lane
        const maxChord = Math.sqrt(Math.max(64, radius * radius - d.perp * d.perp)) * 0.92;
        if (d.prog > maxChord) {
          d.prog = -maxChord;
          d.perp = (Math.random() * 2 - 1) * radius * 0.86;
        }

        const hx = x + perpU * d.perp + dirU * d.prog;
        const hy = y + perpV * d.perp + dirV * d.prog;
        const sLen = intensity.streakLen * d.lenFactor * streakBoost;
        const tx = hx - dirU * sLen;
        const ty = hy - dirV * sLen;

        // Smooth edge alpha inside circular rain cell
        const edgeRatio = Math.min(1, Math.abs(d.prog) / Math.max(1, maxChord));
        const alpha = Math.max(0, 1 - edgeRatio * edgeRatio) * (this.isDark ? 0.92 : 0.86);
        if (alpha < 0.06) continue;

        const grad = ctx.createLinearGradient(tx, ty, hx, hy);
        if (this.isDark) {
          grad.addColorStop(0, 'rgba(125, 211, 252, 0)');
          grad.addColorStop(1, `rgba(96, 165, 250, ${alpha.toFixed(3)})`);
        } else {
          grad.addColorStop(0, 'rgba(56, 189, 248, 0.05)');
          grad.addColorStop(1, `rgba(29, 78, 216, ${alpha.toFixed(3)})`);
        }

        // Tapered raindrop streak tilted in the exact wind_angle direction
        ctx.beginPath();
        ctx.moveTo(tx, ty);
        ctx.lineTo(hx, hy);
        ctx.strokeStyle = grad;
        ctx.lineWidth = intensity.dropWidth;
        ctx.stroke();

        // Crisp teardrop tip
        ctx.beginPath();
        ctx.arc(hx, hy, intensity.dropWidth * 0.72, 0, Math.PI * 2);
        ctx.fillStyle = this.isDark
          ? `rgba(147, 197, 253, ${alpha.toFixed(3)})`
          : `rgba(30, 64, 175, ${alpha.toFixed(3)})`;
        ctx.fill();
      }
    }
  }

  destroy() {
    this.destroyed = true;
    this.stopLoop();
    if (this.map) {
      this.map.off('moveend', this._onMoveEnd);
      this.map.off('zoomstart', this._onZoomStart);
      this.map.off('zoomend', this._onZoomEnd);
      this.map.off('resize', this._onResize);
    }
    if (typeof document !== 'undefined') {
      document.removeEventListener('visibilitychange', this._onVisibility);
    }
    if (this.canvas && this.canvas.parentNode) {
      this.canvas.parentNode.removeChild(this.canvas);
    }
    this.canvas = null;
    this.ctx = null;
    this.pane = null;
    this.map = null;
  }
}
