<script>
  import { i18n } from '$lib/i18n.svelte.js';
</script>

<svelte:head><title>Política de Privacidad · Chiro</title></svelte:head>

<div class="legal-page">
  <a class="back-link" href="/">← Volver</a>
  <h1 class="headline">Política de Privacidad</h1>
  <p class="meta">Última actualización: 5 de agosto de 2026</p>

  <div class="legal-content">
    <h2>1. Información que Recopilamos</h2>
    <p><strong>Información de cuenta:</strong> Nombre, dirección de email y contraseña (almacenada de forma hasheada).</p>
    <p><strong>Datos financieros:</strong> Ingresos, gastos, categorías, presupuestos, préstamos, cuotas, cuentas y transacciones que usted registra voluntariamente.</p>
    <p><strong>Información de uso:</strong> Dirección IP, tipo de navegador, sistema operativo, páginas visitadas y tiempo de uso (solo con su consentimiento para analytics).</p>
    <p><strong>Cookies:</strong> Utilizamos cookies estrictamente necesarias para el funcionamiento del Servicio. Ver sección 8 para detalles.</p>

    <h2>2. Base Legal del Procesamiento (GDPR)</h2>
    <p>Procesamos sus datos bajo las siguientes bases legales:</p>
    <table>
      <thead>
        <tr><th>Actividad</th><th>Base Legal</th><th>Artículo GDPR</th></tr>
      </thead>
      <tbody>
        <tr><td>Creación de cuenta</td><td>Ejecución de contrato</td><td>Art. 6(1)(b)</td></tr>
        <tr><td>Procesamiento de datos financieros</td><td>Ejecución de contrato + consentimiento explícito</td><td>Art. 6(1)(b), Art. 9(2)(a)</td></tr>
        <tr><td>Emails transaccionales</td><td>Ejecución de contrato</td><td>Art. 6(1)(b)</td></tr>
        <tr><td>Mejora del servicio</td><td>Interés legítimo</td><td>Art. 6(1)(f)</td></tr>
        <tr><td>Comunicaciones de marketing</td><td>Consentimiento</td><td>Art. 6(1)(a)</td></tr>
      </tbody>
    </table>

    <h2>3. Cómo Usamos su Información</h2>
    <ul>
      <li>Proveer y mantener el Servicio</li>
      <li>Procesar transacciones y enviar notificaciones relacionadas</li>
      <li>Enviar emails de verificación, reseteo de contraseña y actualizaciones del servicio</li>
      <li>Mejorar la experiencia del usuario y desarrollar nuevas funcionalidades</li>
      <li>Detectar y prevenir fraude, abuso y problemas técnicos</li>
      <li>Cumplir obligaciones legales</li>
    </ul>

    <h2>4. Divulgación a Terceros</h2>
    <p>Compartimos su información únicamente con los siguientes sub-procesadores:</p>
    <table>
      <thead>
        <tr><th>Proveedor</th><th>Propósito</th><th>País</th><th>Protecciones</th></tr>
      </thead>
      <tbody>
        <tr><td>Supabase</td><td>Base de datos</td><td>EE.UU.</td><td>SCCs, cifrado en tránsito y reposo</td></tr>
        <tr><td>Vercel</td><td>Hosting y CDN</td><td>EE.UU.</td><td>SCCs, cifrado en tránsito</td></tr>
        <tr><td>Resend</td><td>Emails transaccionales</td><td>EE.UU.</td><td>SCCs, cifrado</td></tr>
      </tbody>
    </table>
    <p><strong>No vendemos</strong> su información personal a terceros bajo ninguna circunstancia.</p>

    <h2>5. Transferencias Internacionales de Datos</h2>
    <p>Sus datos pueden ser transferidos y procesados en países fuera de Ecuador, incluyendo Estados Unidos. Estas transferencias están protegidas por:</p>
    <ul>
      <li>Cláusulas Contractuales Estándar (SCCs) aprobadas por la Comisión Europea</li>
      <li>Acuerdos de Procesamiento de Datos con todos los sub-procesadores</li>
      <li>Cifrado en tránsito (TLS 1.3) y en reposo (AES-256)</li>
    </ul>
    <p>Puede solicitar una copia de las garantías aplicables contactando privacidad@chiro.app.</p>

    <h2>6. Sus Derechos (GDPR - Arts. 15-22)</h2>
    <p>Usted tiene los siguientes derechos sobre sus datos personales:</p>
    <ul>
      <li><strong>Derecho de acceso:</strong> Solicitar una copia de todos sus datos personales</li>
      <li><strong>Derecho de rectificación:</strong> Corregir datos inexactos o incompletos</li>
      <li><strong>Derecho de supresión:</strong> Solicitar la eliminación de sus datos ("derecho al olvido")</li>
      <li><strong>Derecho a la portabilidad:</strong> Recibir sus datos en formato estructurado y de uso común</li>
      <li><strong>Derecho de oposición:</strong> Oponerse al procesamiento basado en interés legítimo</li>
      <li><strong>Derecho a la restricción:</strong> Solicitar la limitación del procesamiento</li>
      <li><strong>Derecho de revocación:</strong> Retirar el consentimiento en cualquier momento</li>
    </ul>
    <p><strong>Para ejercer sus derechos:</strong> Contacte privacidad@chiro.app. Responderemos dentro de los 30 días hábiles. También puede:</p>
    <ul>
      <li>Exportar datos desde Configuración > Exportar datos</li>
      <li>Eliminar su cuenta desde Configuración > Eliminar cuenta</li>
      <li>Corregir información desde Configuración > Perfil</li>
    </ul>

    <h2>7. Derechos CCPA (California)</h2>
    <p>Si usted es residente de California, tiene derechos adicionales bajo la CCPA/CPRA:</p>
    <ul>
      <li><strong>Derecho a saber:</strong> Solicitar las categorías de información personal recopilada</li>
      <li><strong>Derecho a eliminar:</strong> Solicitar la eliminación de su información personal</li>
      <li><strong>Derecho a opt-out:</strong> Optar por no vender su información personal (no vendemos datos)</li>
      <li><strong>Derecho a no discriminación:</strong> No ser discriminado por ejercer sus derechos</li>
    </ul>
    <p><strong>No vendemos ni compartimos</strong> su información personal con terceros para fines de marketing.</p>

    <h2>8. Cookies y Tecnologías de Rastreo</h2>
    <table>
      <thead>
        <tr><th>Categoría</th><th>Propósito</th><th>Duración</th><th>Consentimiento</th></tr>
      </thead>
      <tbody>
        <tr><td>Estrictamente necesarias</td><td>Autenticación, sesión, seguridad</td><td>Sesión</td><td>No requerido</td></tr>
        <tr><td>Funcionales</td><td>Idioma, preferencias, AMOLED mode</td><td>1 año</td><td>Sí</td></tr>
        <tr><td>Analytics</td><td>Estadísticas de uso (si está habilitado)</td><td>2 años</td><td>Sí</td></tr>
      </tbody>
    </table>
    <p>Puede gestionar sus preferencias de cookies desde la configuración del navegador.</p>

    <h2>9. Seguridad de los Datos</h2>
    <p>Implementamos medidas de seguridad técnicas y organizativas para proteger su información:</p>
    <ul>
      <li>Cifrado en tránsito (TLS 1.3)</li>
      <li>Cifrado en reposo (AES-256)</li>
      <li>Autenticación JWT con expiración</li>
      <li>Rate limiting en endpoints sensibles</li>
      <li>Acceso restringido basado en roles</li>
      <li>Backups regulares</li>
    </ul>

    <h2>10. Notificación de Breach de Datos</h2>
    <p>En caso de un breach de datos personal que afecte su información:</p>
    <ul>
      <li>Notificaremos a la autoridad de protección de datos competente dentro de las 72 horas (GDPR Art. 33)</li>
      <li>Notificaremos a los usuarios afectados sin demora injustificada si el breach implica alto riesgo (Art. 34)</li>
      <li>Documentaremos el breach, sus efectos y las acciones correctivas tomadas</li>
      <li>Le proporcionaremos pasos para protegerse cuando sea aplicable</li>
    </ul>

    <h2>11. Retención de Datos</h2>
    <table>
      <thead>
        <tr><th>Tipo de Dato</th><th>Período de Retención</th><th>Justificación</th></tr>
      </thead>
      <tbody>
        <tr><td>Datos de cuenta</td><td>Duración de la cuenta + 30 días</td><td>Ejecución de contrato</td></tr>
        <tr><td>Datos financieros</td><td>5 años después de la eliminación</td><td>Obligaciones fiscales (SRI Ecuador)</td></tr>
        <tr><td>Logs de uso</td><td>12 meses</td><td>Interés legítimo (seguridad)</td></tr>
        <tr><td>Consentimiento de marketing</td><td>Hasta su revocación</td><td>Consentimiento</td></tr>
        <tr><td>Datos de facturación</td><td>5 años</td><td>Obligaciones contables</td></tr>
      </tbody>
    </table>

    <h2>12. Datos de Menores</h2>
    <p>El Servicio no está dirigido a menores de 16 años. No recopilamos intencionalmente información de menores. Si descubrimos que hemos recopilado datos de un menor, eliminaremos esa información inmediatamente.</p>

    <h2>13. Datos Sensibles</h2>
    <p>Sus datos financieros pueden clasificarse como datos sensibles bajo la legislación ecuatoriana. Procesamos estos datos con su consentimiento explícito y únicamente para los fines descritos en esta política. Usted puede retirar este consentimiento en cualquier momento, lo que puede limitar su uso del Servicio.</p>

    <h2>14. Responsable del Tratamiento</h2>
    <p><strong>Responsable:</strong> Chiro<br />Email: privacidad@chiro.app<br />Dirección: Guayaquil, Ecuador</p>
    <p><strong>Delegado de Protección de Datos:</strong> dpo@chiro.app</p>

    <h2>15. Autoridad de Supervisión</h2>
    <p>Si considera que el procesamiento de sus datos viola la normativa de protección de datos, tiene derecho a presentar una queja ante:</p>
    <ul>
      <li><strong>Ecuador:</strong> Superintendencia de Protección de Datos Personales</li>
      <li><strong>UE:</strong> Autoridad de supervisión de su país de residencia</li>
    </ul>

    <h2>16. Cambios en esta Política</h2>
    <p>Podemos actualizar esta política periódicamente. Los cambios materiales serán notificados por email con 30 días de anticipación. La versión actualizada estará disponible en esta página con la fecha de "Última actualización"修订.</p>

    <h2>17. Contacto</h2>
    <p>Para consultas sobre privacidad:</p>
    <p>Email: privacidad@chiro.app<br />DPO: dpo@chiro.app<br />Dirección: Guayaquil, Ecuador</p>
  </div>
</div>
