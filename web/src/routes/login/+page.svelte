<script>
  import { goto } from '$app/navigation';
  import { i18n } from '$lib/i18n.svelte.js';
  import { login, register, S } from '$lib/stores.svelte.js';
  import Button from '$lib/components/ui/button.svelte';
  import Input from '$lib/components/ui/input.svelte';
  import Label from '$lib/components/ui/label.svelte';
  import Card from '$lib/components/ui/card.svelte';

  let mode = $state('login');
  let name = $state('');
  let email = $state('');
  let password = $state('');
  let err = $state('');
  let termsAccepted = $state(false);
  let privacyAccepted = $state(false);

  async function submit() {
    err = '';
    if (mode === 'register' && !name.trim()) {
      err = i18n.t('auth.nameRequired');
      return;
    }
    if (!email.trim()) {
      err = i18n.t('auth.emailRequired');
      return;
    }
    if (password.length < 6) {
      err = i18n.t('auth.passwordRequired');
      return;
    }
    if (mode === 'register' && (!termsAccepted || !privacyAccepted)) {
      err = 'Debes aceptar los Términos y la Política de Privacidad';
      return;
    }
    try {
      if (mode === 'login') await login(email, password);
      else await register(name, email, password, termsAccepted, privacyAccepted);
      goto('/dashboard');
    } catch {
      err = mode === 'login' ? i18n.t('auth.loginError') : i18n.t('auth.registerError');
    }
  }
</script>

<svelte:head><title>{mode === 'login' ? i18n.t('auth.login') : i18n.t('auth.register')} · Chiro</title></svelte:head>

<div class="min-h-screen flex items-center justify-center p-4">
  <Card class="w-full max-w-sm p-8">
    <div class="text-center mb-6">
      <div class="w-12 h-12 rounded-xl bg-primary/10 border border-primary/20 text-primary font-black text-xl mx-auto mb-3 flex items-center justify-center">
        C
      </div>
      <h1 class="text-xl font-bold">{mode === 'login' ? i18n.t('auth.loginTitle') : 'Crear cuenta'}</h1>
      <p class="text-sm text-muted-foreground mt-1">{i18n.t('auth.loginSub')}</p>
    </div>

    <div class="flex gap-2 justify-center mb-6">
      <Button variant={i18n.lang === 'es' ? 'default' : 'outline'} size="sm" onclick={() => i18n.setLang('es')}>ES</Button>
      <Button variant={i18n.lang === 'en' ? 'default' : 'outline'} size="sm" onclick={() => i18n.setLang('en')}>EN</Button>
    </div>

    <form onsubmit={(e) => { e.preventDefault(); submit(); }} class="space-y-4">
      {#if mode === 'register'}
        <div class="space-y-2">
          <Label for="login-name">{i18n.t('auth.name')}</Label>
          <Input id="login-name" bind:value={name} placeholder="Chiro" autocomplete="name" />
        </div>
      {/if}
      <div class="space-y-2">
        <Label for="login-email">{i18n.t('auth.email')}</Label>
        <Input id="login-email" bind:value={email} type="email" placeholder={i18n.t('auth.emailPlaceholder')} autocomplete="email" />
      </div>
      <div class="space-y-2">
        <Label for="login-password">{i18n.t('auth.password')}</Label>
        <Input id="login-password" bind:value={password} type="password" placeholder={i18n.t('auth.passwordPlaceholder')} autocomplete={mode === 'login' ? 'current-password' : 'new-password'} />
      </div>

      {#if mode === 'register'}
        <div class="space-y-2">
          <label class="flex items-start gap-2 cursor-pointer">
            <input type="checkbox" bind:checked={termsAccepted} class="mt-0.5" />
            <span class="text-sm text-muted-foreground">Acepto los <a href="/legal/tos" target="_blank" class="text-primary hover:underline">Términos de Servicio</a></span>
          </label>
          <label class="flex items-start gap-2 cursor-pointer">
            <input type="checkbox" bind:checked={privacyAccepted} class="mt-0.5" />
            <span class="text-sm text-muted-foreground">Acepto la <a href="/legal/privacy" target="_blank" class="text-primary hover:underline">Política de Privacidad</a></span>
          </label>
        </div>
      {/if}

      {#if err}
        <p class="text-sm text-destructive">{err}</p>
      {/if}

      <Button type="submit" class="w-full" disabled={S.busy}>
        {mode === 'login' ? i18n.t('auth.login') : i18n.t('auth.register')}
      </Button>
    </form>

    <Button variant="ghost" class="w-full mt-4" onclick={() => (mode = mode === 'login' ? 'register' : 'login')}>
      {mode === 'login' ? i18n.t('auth.register') : i18n.t('auth.login')}
    </Button>
  </Card>
</div>
