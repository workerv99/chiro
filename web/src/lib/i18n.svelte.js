// i18n es/en con auto-detección del navegador (port del app Android).
export const messages = {
  es: {
    tabs: { expenses: 'Gastos', stats: 'Estadísticas', budgets: 'Presupuestos', loans: 'Préstamos', config: 'Ajustes' },
    common: {
      save: 'Guardar', cancel: 'Cancelar', delete: 'Eliminar', add: 'Añadir', loading: 'Cargando…',
      logout: 'Cerrar sesión', pay: 'Pagar', skip: 'Saltar', expense: 'Gasto', income: 'Ingreso',
      transfer: 'Transferencia', required: 'Completa todos los campos'
    },
    summary: { income: 'Ingresos', expense: 'Gastos', balance: 'Balance' },
    expenses: {
      title: 'Gastos', newExpense: 'Nuevo gasto', editExpense: 'Editar gasto',
      description: 'Descripción', descriptionPlaceholder: 'Ej. Supermercado',
      amount: 'Importe', amountPlaceholder: '0.00', date: 'Fecha', datePlaceholder: 'DD/MM/AAAA',
      category: 'Categoría', noCategory: 'Sin categoría', account: 'Cuenta', noAccount: 'Sin cuenta',
      accountFrom: 'Desde', accountTo: 'Hacia', transferHint: 'Movimiento entre cuentas, no afecta el balance.',
      notes: 'Notas', notesPlaceholder: 'Opcional', tags: 'Etiquetas',
      descriptionRequired: 'Escribe una descripción', amountRequired: 'Importe inválido',
      dateRequired: 'Fecha inválida (DD/MM/AAAA)', accountRequired: 'Elige ambas cuentas',
      createError: 'No se pudo guardar', empty: 'Sin movimientos este mes',
      deleteConfirm: '¿Eliminar este movimiento?', deleteTransferConfirm: '¿Eliminar esta transferencia?'
    },
    stats: { title: 'Estadísticas', byCategory: 'Por categoría', monthly: 'Ingresos vs gastos', summary: 'Resumen', outstanding: 'Pendiente de préstamos', empty: 'Sin datos este mes' },
    budgets: {
      title: 'Presupuestos', newBudget: 'Nuevo presupuesto', editBudget: 'Editar presupuesto',
      amount: 'Importe', category: 'Categoría', allCategories: 'Todas las categorías',
      exceeded: 'Excedido', empty: 'Sin presupuestos', deleteConfirm: '¿Eliminar presupuesto?'
    },
    loans: {
      title: 'Préstamos', newLoan: 'Nuevo préstamo',
      person: 'Persona', personRequired: 'Elige la persona', principal: 'Capital', rate: 'Tasa %',
      interestSimple: 'Interés simple', interestCompound: 'Interés compuesto',
      installments: 'Cuotas', installmentsRequired: 'Indica el número de cuotas', frequency: 'Frecuencia',
      weekly: 'Semanal', biweekly: 'Quincenal', monthly: 'Mensual', startDate: 'Fecha inicial',
      dueDate: 'Primera cuota', description: 'Descripción', amountRequired: 'Capital inválido', remaining: 'Pendiente',
      paidLabel: 'Pagado', progress: 'Cuotas', paid: 'Pagado', pending: 'Pendiente', interest: 'Interés', partial: 'Parcial',
      schedule: 'Calendario', installmentN: 'Cuota', paymentAmount: 'Importe', paymentDate: 'Fecha',
      payNext: 'Pagar siguiente', payNow: 'Pagar ahora', cascade: 'Reestructurar', unpay: 'Deshacer pago',
      empty: 'Sin préstamos', deleteConfirm: '¿Eliminar préstamo y su calendario?'
    },
    auth: {
      loginTitle: 'Bienvenido', loginSub: 'Controla tus gastos, hoy.',
      login: 'Entrar', register: 'Crear cuenta', name: 'Nombre', email: 'Email',
      emailPlaceholder: 'tu@email.com', password: 'Contraseña', passwordPlaceholder: 'Mínimo 6 caracteres',
      nameRequired: 'Escribe tu nombre', emailRequired: 'Escribe tu email',
      passwordRequired: 'Mínimo 6 caracteres', loginError: 'Email o contraseña incorrectos',
      registerError: 'No se pudo crear la cuenta'
    },
    config: {
      title: 'Ajustes', accounts: 'Cuentas', categories: 'Categorías', persons: 'Personas',
      tags: 'Etiquetas', piggy: 'Alcancías', bills: 'Facturas', name: 'Nombre', currency: 'Moneda',
      type: 'Tipo', color: 'Color', notes: 'Notas', target: 'Objetivo', current: 'Actual', amount: 'Importe',
      nextDate: 'Próxima fecha', frequency: 'Frecuencia', day: 'Día',
      dueBills: 'Facturas por pagar', noDueBills: 'Nada pendiente', editItem: 'Editar',
      deleteConfirm: '¿Eliminar?', empty: 'Sin elementos', session: 'Sesión', loggedAs: 'Conectado como'
    }
  },
  en: {
    tabs: { expenses: 'Expenses', stats: 'Stats', budgets: 'Budgets', loans: 'Loans', config: 'Settings' },
    common: {
      save: 'Save', cancel: 'Cancel', delete: 'Delete', add: 'Add', loading: 'Loading…',
      logout: 'Log out', pay: 'Pay', skip: 'Skip', expense: 'Expense', income: 'Income',
      transfer: 'Transfer', required: 'Fill in all fields'
    },
    summary: { income: 'Income', expense: 'Expenses', balance: 'Balance' },
    expenses: {
      title: 'Expenses', newExpense: 'New expense', editExpense: 'Edit expense',
      description: 'Description', descriptionPlaceholder: 'e.g. Groceries',
      amount: 'Amount', amountPlaceholder: '0.00', date: 'Date', datePlaceholder: 'DD/MM/YYYY',
      category: 'Category', noCategory: 'No category', account: 'Account', noAccount: 'No account',
      accountFrom: 'From', accountTo: 'To', transferHint: 'Moves money between accounts, no balance impact.',
      notes: 'Notes', notesPlaceholder: 'Optional', tags: 'Tags',
      descriptionRequired: 'Enter a description', amountRequired: 'Invalid amount',
      dateRequired: 'Invalid date (DD/MM/YYYY)', accountRequired: 'Pick both accounts',
      createError: 'Could not save', empty: 'No transactions this month',
      deleteConfirm: 'Delete this transaction?', deleteTransferConfirm: 'Delete this transfer?'
    },
    stats: { title: 'Stats', byCategory: 'By category', monthly: 'Income vs expenses', summary: 'Summary', outstanding: 'Outstanding loans', empty: 'No data this month' },
    budgets: {
      title: 'Budgets', newBudget: 'New budget', editBudget: 'Edit budget',
      amount: 'Amount', category: 'Category', allCategories: 'All categories',
      exceeded: 'Exceeded', empty: 'No budgets yet', deleteConfirm: 'Delete budget?'
    },
    loans: {
      title: 'Loans', newLoan: 'New loan',
      person: 'Person', personRequired: 'Pick a person', principal: 'Principal', rate: 'Rate %',
      interestSimple: 'Simple interest', interestCompound: 'Compound interest',
      installments: 'Installments', installmentsRequired: 'Number of installments required', frequency: 'Frequency',
      weekly: 'Weekly', biweekly: 'Biweekly', monthly: 'Monthly', startDate: 'Start date',
      dueDate: 'First due date', description: 'Description', amountRequired: 'Invalid principal', remaining: 'Remaining',
      paidLabel: 'Paid', progress: 'Installments', paid: 'Paid', pending: 'Pending', interest: 'Interest', partial: 'Partial',
      schedule: 'Schedule', installmentN: 'Installment', paymentAmount: 'Amount', paymentDate: 'Date',
      payNext: 'Pay next', payNow: 'Pay now', cascade: 'Restructure', unpay: 'Undo payment',
      empty: 'No loans yet', deleteConfirm: 'Delete loan and its schedule?'
    },
    auth: {
      loginTitle: 'Welcome back', loginSub: 'Track your money, today.',
      login: 'Log in', register: 'Create account', name: 'Name', email: 'Email',
      emailPlaceholder: 'you@email.com', password: 'Password', passwordPlaceholder: 'Min 6 characters',
      nameRequired: 'Enter your name', emailRequired: 'Enter your email',
      passwordRequired: 'Min 6 characters', loginError: 'Wrong email or password',
      registerError: 'Could not create account'
    },
    config: {
      title: 'Settings', accounts: 'Accounts', categories: 'Categories', persons: 'People',
      tags: 'Tags', piggy: 'Piggy banks', bills: 'Bills', name: 'Name', currency: 'Currency',
      type: 'Type', color: 'Color', notes: 'Notes', target: 'Target', current: 'Current', amount: 'Amount',
      nextDate: 'Next date', frequency: 'Frequency', day: 'Day',
      dueBills: 'Bills due', noDueBills: 'Nothing due', editItem: 'Edit',
      deleteConfirm: 'Delete?', empty: 'Nothing here yet', session: 'Session', loggedAs: 'Signed in as'
    }
  }
};

function detect() {
  try {
    const l = (navigator.language || 'en').toLowerCase();
    return l.startsWith('es') ? 'es' : 'en';
  } catch {
    return 'en';
  }
}

let initial = detect();
try {
  const saved = localStorage.getItem('chiro_lang');
  if (saved === 'es' || saved === 'en') initial = saved;
} catch { /* ignore */ }

function resolve(obj, path) {
  return path.split('.').reduce((o, k) => (o == null ? undefined : o[k]), obj);
}

export const i18n = $state({
  lang: initial,
  setLang(l) {
    i18n.lang = l;
    try {
      localStorage.setItem('chiro_lang', l);
    } catch { /* ignore */ }
  },
  get t() {
    const m = messages[i18n.lang] || messages.en;
    return (key) => resolve(m, key) || key;
  }
});
