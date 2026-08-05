<script>
  import { onMount } from 'svelte';

  let showBanner = $state(false);
  let analyticsConsent = $state(false);
  let functionalConsent = $state(true);

  onMount(() => {
    const consent = localStorage.getItem('chiro_cookie_consent');
    if (!consent) {
      showBanner = true;
    }
  });

  function acceptAll() {
    analyticsConsent = true;
    functionalConsent = true;
    saveConsent();
  }

  function acceptSelected() {
    saveConsent();
  }

  function rejectAll() {
    analyticsConsent = false;
    functionalConsent = true;
    saveConsent();
  }

  function saveConsent() {
    localStorage.setItem('chiro_cookie_consent', JSON.stringify({
      analytics: analyticsConsent,
      functional: functionalConsent,
      timestamp: new Date().toISOString()
    }));
    showBanner = false;
  }
</script>

{#if showBanner}
  <div class="cookie-banner">
    <div class="cookie-content">
      <div class="cookie-text">
        <h3>Usamos cookies</h3>
        <p>Utilizamos cookies para mejorar tu experiencia. Las cookies estrictamente necesarias son siempre activas. Puedes configurar las opcionales.</p>
      </div>
      <div class="cookie-actions">
        <button class="btn btn-small" onclick={rejectAll}>Rechazar opcionales</button>
        <button class="btn btn-small" onclick={acceptSelected}>Aceptar seleccionadas</button>
        <button class="btn btn-small btn-primary" onclick={acceptAll}>Aceptar todas</button>
      </div>
    </div>
    <div class="cookie-options">
      <label class="cookie-option">
        <input type="checkbox" checked disabled />
        <span>Estrictamente necesarias (siempre activas)</span>
      </label>
      <label class="cookie-option">
        <input type="checkbox" bind:checked={functionalConsent} />
        <span>Funcionales (idioma, preferencias)</span>
      </label>
      <label class="cookie-option">
        <input type="checkbox" bind:checked={analyticsConsent} />
        <span>Analytics (estadísticas de uso)</span>
      </label>
    </div>
    <a href="/legal/privacy" class="cookie-link" target="_blank">Más información</a>
  </div>
{/if}

<style>
  .cookie-banner {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    background: var(--surface);
    border-top: 1px solid var(--border);
    padding: 16px 24px;
    z-index: 100;
    box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.3);
  }

  .cookie-content {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    max-width: 960px;
    margin: 0 auto;
  }

  .cookie-text h3 {
    font-size: 0.95rem;
    font-weight: 700;
    margin-bottom: 4px;
  }

  .cookie-text p {
    color: var(--ink-dim);
    font-size: 0.82rem;
    margin: 0;
  }

  .cookie-actions {
    display: flex;
    gap: 8px;
    flex-shrink: 0;
  }

  .cookie-options {
    display: flex;
    gap: 16px;
    max-width: 960px;
    margin: 12px auto 0;
    padding-top: 12px;
    border-top: 1px solid var(--border);
  }

  .cookie-option {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 0.78rem;
    color: var(--ink-dim);
    cursor: pointer;
  }

  .cookie-option input {
    width: 14px;
    height: 14px;
    accent-color: var(--indigo);
  }

  .cookie-link {
    display: block;
    text-align: center;
    color: var(--indigo);
    font-size: 0.75rem;
    margin-top: 8px;
    text-decoration: underline;
  }

  @media (max-width: 640px) {
    .cookie-content {
      flex-direction: column;
      text-align: center;
    }
    .cookie-options {
      flex-direction: column;
      align-items: center;
    }
  }
</style>
