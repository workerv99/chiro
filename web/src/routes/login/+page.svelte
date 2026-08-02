<script>
  import { goto } from '$app/navigation';
  import { i18n } from '$lib/i18n.svelte.js';
  import { login, register, S } from '$lib/stores.svelte.js';

  let mode = $state('login');
  let name = $state('');
  let email = $state('');
  let password = $state('');
  let err = $state('');

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
    try {
      if (mode === 'login') await login(email, password);
      else await register(name, email, password);
      goto('/');
    } catch {
      err = mode === 'login' ? i18n.t('auth.loginError') : i18n.t('auth.registerError');
    }
  }
</script>

<svelte:head><title>{mode === 'login' ? i18n.t('auth.login') : i18n.t('auth.register')} · Chiro</title></svelte:head>

<div class="login-wrap">
  <div class="login-card">
    <div class="login-brand">
      <div class="login-mark">C</div>
      <h1 class="headline">{i18n.t('auth.loginTitle')}</h1>
      <p class="meta">{i18n.t('auth.loginSub')}</p>
    </div>

    <div class="login-lang">
      <button class="tag-chip" class:active={i18n.lang === 'es'} onclick={() => i18n.setLang('es')}>ES</button>
      <button class="tag-chip" class:active={i18n.lang === 'en'} onclick={() => i18n.setLang('en')}>EN</button>
    </div>

    <form onsubmit={(e) => { e.preventDefault(); submit(); }}>
      {#if mode === 'register'}
        <div class="form-field">
          <label for="login-name">{i18n.t('auth.name')}</label>
          <input id="login-name" bind:value={name} placeholder="Chiro" autocomplete="name" />
        </div>
      {/if}
      <div class="form-field">
        <label for="login-email">{i18n.t('auth.email')}</label>
        <input id="login-email" bind:value={email} type="email" placeholder={i18n.t('auth.emailPlaceholder')} autocomplete="email" />
      </div>
      <div class="form-field">
        <label for="login-password">{i18n.t('auth.password')}</label>
        <input id="login-password" bind:value={password} type="password" placeholder={i18n.t('auth.passwordPlaceholder')} autocomplete={mode === 'login' ? 'current-password' : 'new-password'} />
      </div>
      {#if err}
        <p class="error-text" role="alert">{err}</p>
      {/if}
      <button class="btn btn-primary login-submit" type="submit" disabled={S.busy}>
        {mode === 'login' ? i18n.t('auth.login') : i18n.t('auth.register')}
      </button>
    </form>

    <button class="btn btn-cancel" onclick={() => (mode = mode === 'login' ? 'register' : 'login')}>
      {mode === 'login' ? i18n.t('auth.register') : i18n.t('auth.login')}
    </button>
  </div>
</div>

<style>
  .login-wrap {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 16px;
  }
  .login-card {
    width: 100%;
    max-width: 400px;
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: 18px;
    padding: 28px 20px;
  }
  .login-brand {
    text-align: center;
    margin-bottom: 16px;
  }
  .login-brand p {
    margin: 4px 0 0;
  }
  .login-mark {
    width: 48px;
    height: 48px;
    margin: 0 auto 10px;
    border-radius: 12px;
    background: var(--indigo-tint);
    border: 1px solid rgba(91, 124, 246, 0.27);
    color: var(--indigo);
    font-weight: 900;
    font-size: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .login-lang {
    display: flex;
    gap: 8px;
    justify-content: center;
    margin-bottom: 16px;
  }
  .login-lang .tag-chip.active {
    color: var(--indigo);
    border-color: var(--indigo);
    background: var(--indigo-tint);
  }
  .login-submit {
    width: 100%;
  }
</style>
