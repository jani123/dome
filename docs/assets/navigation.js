document.documentElement.classList.add('js');
const theme = document.getElementById('theme-toggle');
const preference = matchMedia('(prefers-color-scheme: dark)');
function currentTheme() { return document.documentElement.dataset.theme || (preference.matches ? 'dark' : 'light'); }
function labelTheme() { theme.textContent = currentTheme() === 'dark' ? 'Light' : 'Dark'; theme.setAttribute('aria-label', `Switch to ${currentTheme() === 'dark' ? 'light' : 'dark'} theme`); }
theme.hidden = false;
labelTheme();
preference.addEventListener('change', labelTheme);
theme.addEventListener('click', () => {
  const next = currentTheme() === 'dark' ? 'light' : 'dark';
  document.documentElement.dataset.theme = next;
  try { localStorage.setItem('dome-theme', next); } catch {}
  labelTheme();
});
const menu = document.getElementById('menu-toggle');
const sidebar = document.getElementById('sidebar');
menu.hidden = false;
const background = [document.getElementById('main'), document.querySelector('.page-outline')];
function closeMenu() {
  background.forEach(element => { if (element) element.inert = false; });
  sidebar.classList.remove('is-open');
  document.body.classList.remove('menu-open');
  menu.setAttribute('aria-expanded', 'false');
  menu.textContent = 'Menu';
}
menu.addEventListener('click', () => {
  if (sidebar.classList.contains('is-open')) { closeMenu(); return; }
  sidebar.classList.add('is-open');
  background.forEach(element => { if (element) element.inert = true; });
  document.body.classList.add('menu-open');
  menu.setAttribute('aria-expanded', 'true');
  menu.textContent = 'Close';
});
document.addEventListener('keydown', event => {
  if (!sidebar.classList.contains('is-open')) return;
  if (event.key === 'Escape') { closeMenu(); menu.focus(); }
  if (event.key === 'Tab') {
    const targets = [menu, ...sidebar.querySelectorAll('a[href]')];
    const first = targets[0], last = targets[targets.length - 1];
    if (event.shiftKey && document.activeElement === first) { event.preventDefault(); last.focus(); }
    else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first.focus(); }
  }
});
matchMedia('(max-width: 800px)').addEventListener('change', closeMenu);
if (matchMedia('(max-width: 800px)').matches) document.querySelector('.page-outline details')?.removeAttribute('open');
// Follow the actual header height when text zoom or a long version wraps.
new ResizeObserver(entries => {
  document.documentElement.style.setProperty('--header-height', `${entries[0].target.getBoundingClientRect().height}px`);
}).observe(document.querySelector('.header'));
const versions = document.getElementById('version-select');
versions.addEventListener('change', () => {
  window.location.assign(versions.value);
});
