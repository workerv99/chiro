<script>
  import { i18n } from '$lib/i18n.svelte.js';
  import { A, loadToken } from '$lib/api.svelte.js';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { DollarSign, Hash, Percent, Users, FileText, Shield } from 'lucide-svelte';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';

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
  <div class="min-h-screen">
    <nav class="flex items-center justify-between px-6 py-4 border-b">
      <div class="flex items-center gap-2">
        <div class="w-8 h-8 rounded-lg bg-primary/10 border border-primary/20 text-primary font-bold text-sm flex items-center justify-center">C</div>
        <span class="font-bold text-lg">Chiro</span>
      </div>
      <div class="flex items-center gap-6">
        <a href="#features" class="text-sm text-muted-foreground hover:text-foreground">Características</a>
        <a href="#pricing" class="text-sm text-muted-foreground hover:text-foreground">Precios</a>
        <a href="/login" class="inline-flex items-center justify-center h-9 px-4 rounded-md bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90">Iniciar sesión</a>
      </div>
    </nav>

    <section class="text-center py-20 px-4 max-w-2xl mx-auto">
      <h1 class="text-4xl md:text-5xl font-extrabold tracking-tight mb-4">
        Controla tu dinero<br />
        <span class="text-primary">sin complicaciones</span>
      </h1>
      <p class="text-lg text-muted-foreground mb-8">Gastos, presupuestos y préstamos en una sola app. Gratis para empezar.</p>
      <div class="flex gap-4 justify-center">
        <a href="/login"><Button size="lg">Empezar gratis</Button></a>
        <a href="#features"><Button variant="outline" size="lg">Ver características</Button></a>
      </div>
    </section>

    <section id="features" class="py-16 px-4 max-w-5xl mx-auto">
      <h2 class="text-3xl font-extrabold text-center mb-2">Todo lo que necesitas</h2>
      <p class="text-muted-foreground text-center mb-10">Herramientas simples para tomar el control de tus finanzas.</p>
      <div class="grid md:grid-cols-3 gap-6">
        {#each features as f}
          <Card class="p-6">
            <div class="w-10 h-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center mb-4">
              <svelte:component this={f.icon} size={20} />
            </div>
            <h3 class="font-bold mb-1">{f.title}</h3>
            <p class="text-sm text-muted-foreground">{f.desc}</p>
          </Card>
        {/each}
      </div>
    </section>

    <section class="py-16 px-4 bg-muted/30">
      <div class="max-w-5xl mx-auto">
        <h2 class="text-3xl font-extrabold text-center mb-10">¿Por qué Chiro?</h2>
        <div class="grid grid-cols-2 md:grid-cols-4 gap-8 max-w-2xl mx-auto">
          <div class="text-center">
            <div class="text-3xl font-extrabold text-primary">3</div>
            <div class="text-sm text-muted-foreground mt-1">minutos para empezar</div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-extrabold text-primary">100%</div>
            <div class="text-sm text-muted-foreground mt-1">gratis para empezar</div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-extrabold text-primary">0</div>
            <div class="text-sm text-muted-foreground mt-1">publicidad en la app</div>
          </div>
          <div class="text-center">
            <div class="text-3xl font-extrabold text-primary">24/7</div>
            <div class="text-sm text-muted-foreground mt-1">acceso desde cualquier lado</div>
          </div>
        </div>
      </div>
    </section>

    <section id="pricing" class="py-16 px-4 max-w-3xl mx-auto">
      <h2 class="text-3xl font-extrabold text-center mb-2">Planes simples</h2>
      <p class="text-muted-foreground text-center mb-10">Empieza gratis, upgrade cuando lo necesites.</p>
      <div class="grid md:grid-cols-2 gap-6">
        {#each plans as plan}
          <Card class="p-8 text-center {plan.primary ? 'border-primary bg-gradient-to-b from-primary/5 to-transparent' : ''}">
            <p class="text-sm font-bold text-muted-foreground uppercase tracking-wider mb-2">{plan.name}</p>
            <p class="text-4xl font-extrabold mb-1">{plan.price}</p>
            <p class="text-sm text-muted-foreground mb-6">{plan.period}</p>
            <ul class="text-sm text-left space-y-3 mb-8">
              {#each plan.features as feat}
                <li class="flex items-center gap-2"><span class="text-primary">✓</span> {feat}</li>
              {/each}
            </ul>
            <a href="/login">
              <Button variant={plan.primary ? 'default' : 'outline'} class="w-full">{plan.cta}</Button>
            </a>
          </Card>
        {/each}
      </div>
    </section>

    <footer class="text-center py-8 px-4 border-t text-sm text-muted-foreground">
      <p>&copy; 2026 Chiro. Todos los derechos reservados.</p>
      <p class="mt-2">
        <a href="/legal/tos" class="hover:text-foreground">Términos</a> ·
        <a href="/legal/privacy" class="hover:text-foreground">Privacidad</a>
      </p>
    </footer>
  </div>
{/if}
