<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import { A, loadToken } from '$lib/api.svelte.js';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { DollarSign, Hash, Percent, Users, FileText, Shield, ChevronLeft, ChevronRight } from 'lucide-svelte';

  let ready = $state(false);

  onMount(() => {
    loadToken();
    if (A.token) goto('/dashboard');
    else ready = true;
  });

  const features = [
    { icon: DollarSign, title: 'Control de gastos', desc: 'Registra ingresos y gastos diarios con categorías y etiquetas.' },
    { icon: Hash, title: 'Presupuestos', desc: 'Establece límites por categoría y monitorea tu progreso.' },
    { icon: Percent, title: 'Préstamos', desc: 'Gestiona préstamos con calendario de cuotas y seguimiento de pagos.' },
    { icon: Users, title: 'Multi-usuario', desc: 'Comparte con tu pareja o familia con cuentas separadas.' },
    { icon: FileText, title: 'Reportes PDF', desc: 'Genera reportes profesionales de préstamos y gastos.' },
    { icon: Shield, title: 'Datos seguros', desc: 'Tus datos están en la nube con encriptación de extremo a extremo.' }
  ];

  const plans = [
    {
      name: 'Free',
      price: '$0',
      period: 'para siempre',
      features: ['50 gastos/mes', '3 cuentas', '10 préstamos', 'Reportes básicos'],
      cta: 'Empezar gratis',
      primary: false
    },
    {
      name: 'Pro',
      price: '$4.99',
      period: '/mes',
      features: ['Gastos ilimitados', 'Cuentas ilimitadas', 'Préstamos ilimitados', 'Reportes PDF', 'Soporte prioritario', 'Exportar datos'],
      cta: 'Suscribirse',
      primary: true
    }
  ];
</script>

<svelte:head>
  <title>Chiro - Controla tu dinero</title>
  <meta name="description" content="Aplicación gratuita para controlar gastos, presupuestos y préstamos." />
</svelte:head>

{#if ready}
  <div class="landing">
    <nav class="landing-nav">
      <div class="landing-brand">
        <div class="brand-mark">C</div>
        <span class="brand-name">Chiro</span>
      </div>
      <div class="landing-nav-links">
        <a href="#features">Características</a>
        <a href="#pricing">Precios</a>
        <a href="/login" class="btn btn-primary btn-small">Iniciar sesión</a>
      </div>
    </nav>

    <section class="hero">
      <h1>Controla tu dinero<br /><span class="hero-highlight">sin complicaciones</span></h1>
      <p class="hero-sub">Gastos, presupuestos y préstamos en una sola app. Gratis para empezar.</p>
      <div class="hero-cta">
        <a href="/login" class="btn btn-primary">Empezar gratis</a>
        <a href="#features" class="btn">Ver características</a>
      </div>
    </section>

    <section id="features" class="section">
      <h2 class="section-title">Todo lo que necesitas</h2>
      <p class="section-sub">Herramientas simples para tomar el control de tus finanzas.</p>
      <div class="features-grid">
        {#each features as f}
          <div class="feature-card">
            <div class="feature-icon">
              <svelte:component this={f.icon} size={20} />
            </div>
            <h3>{f.title}</h3>
            <p>{f.desc}</p>
          </div>
        {/each}
      </div>
    </section>

    <section class="section" style="background:var(--surface)">
      <h2 class="section-title">¿Por qué Chiro?</h2>
      <div class="why-grid">
        <div class="why-item">
          <div class="why-num">3</div>
          <div class="why-label">minutos para empezar</div>
        </div>
        <div class="why-item">
          <div class="why-num">100%</div>
          <div class="why-label">gratis para empezar</div>
        </div>
        <div class="why-item">
          <div class="why-num">0</div>
          <div class="why-label">publicidad en la app</div>
        </div>
        <div class="why-item">
          <div class="why-num">24/7</div>
          <div class="why-label">acceso desde cualquier lado</div>
        </div>
      </div>
    </section>

    <section id="pricing" class="section">
      <h2 class="section-title">Planes simples</h2>
      <p class="section-sub">Empieza gratis, upgrade cuando lo necesites.</p>
      <div class="pricing-grid">
        {#each plans as plan}
          <div class="pricing-card" class:primary={plan.primary}>
            <div class="pricing-name">{plan.name}</div>
            <div class="pricing-price">{plan.price}</div>
            <div class="pricing-period">{plan.period}</div>
            <ul class="pricing-features">
              {#each plan.features as feat}
                <li>{feat}</li>
              {/each}
            </ul>
            <a href="/login" class="btn" class:btn-primary={plan.primary} style="width:100%">{plan.cta}</a>
          </div>
        {/each}
      </div>
    </section>

    <footer class="landing-footer">
      <p>&copy; 2026 Chiro. Todos los derechos reservados.</p>
      <p style="margin-top:8px">
        <a href="/legal/tos" style="color:var(--ink-dim)">Términos</a> ·
        <a href="/legal/privacy" style="color:var(--ink-dim)">Privacidad</a>
      </p>
    </footer>
  </div>
{/if}
