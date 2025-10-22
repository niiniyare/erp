/**
* ERP System - Optimized JavaScript Bundle
* Includes: Alpine.js stores + essential utilities
* Target: <50KB total size
*/
// === ALPINE.JS STORES ===
// App Store
document.addEventListener('alpine:init', () => {
Alpine.store('app', {
sidebarOpen: false,
mobileMenuOpen: false,
theme: 'light',
loading: false,
currentTenant: null,
currentPage: '',
breadcrumbs: [],
activeModal: null,
modalData: null,
overlayVisible: false,
showNotifications: false,
_pendingRequests: 0,
_previousOverflow: '',
_resizeHandler: null,
_keydownHandler: null,
_htmxBeforeHandler: null,
_htmxAfterHandler: null,
get isLoading() {
return this.loading || this._pendingRequests > 0;
},
_persistState(key, value) {
try {
localStorage.setItem(key, typeof value === 'string' ? value : JSON.stringify(value));
return true;
} catch (e) {
console.error(`Failed to persist ${key}:`, e);
return false;
}
},
_loadState(key, defaultValue = null) {
try {
const saved = localStorage.getItem(key);
if (!saved) return defaultValue;
try {
return JSON.parse(saved);
} catch {
return saved;
}
} catch (e) {
console.error(`Failed to load ${key}:`, e);
return defaultValue;
}
},
toggleSidebar() {
this.sidebarOpen = !this.sidebarOpen;
},
closeSidebar() {
this.sidebarOpen = false;
},
toggleMobileMenu() {
this.mobileMenuOpen = !this.mobileMenuOpen;
},
closeMobileMenu() {
this.mobileMenuOpen = false;
},
toggleTheme() {
this.theme = this.theme === 'light' ? 'dark' : 'light';
this.persistTheme();
},
persistTheme() {
this._persistState('theme', this.theme);
document.documentElement.classList.toggle('dark', this.theme === 'dark');
document.documentElement.setAttribute('data-theme', this.theme);
},
loadTheme() {
const saved = this._loadState('theme');
if (saved && (saved === 'light' || saved === 'dark')) {
this.theme = saved;
} else {
this.theme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}
this.persistTheme();
},
setLoading(state) {
this.loading = Boolean(state);
},
setCurrentPage(page) {
this.currentPage = page || '';
},
setBreadcrumbs(crumbs) {
if (!Array.isArray(crumbs)) {
this.breadcrumbs = [];
return;
}
this.breadcrumbs = crumbs.filter(crumb => {
return crumb && (
typeof crumb === 'string' ||
(typeof crumb === 'object' && (crumb.label || crumb.title || crumb.name))
);
});
},
openModal(modalId, data = null) {
if (!modalId) {
console.warn('openModal called without modalId');
return;
}
this.activeModal = modalId;
this.modalData = data;
this.overlayVisible = true;
this._previousOverflow = document.body.style.overflow || '';
document.body.style.overflow = 'hidden';
},
closeModal() {
this.activeModal = null;
this.modalData = null;
this.overlayVisible = false;
document.body.style.overflow = this._previousOverflow;
this._previousOverflow = '';
},
toggleNotifications() {
this.showNotifications = !this.showNotifications;
},
closeNotifications() {
this.showNotifications = false;
},
setTenant(tenant) {
if (!tenant || typeof tenant !== 'object') {
console.warn('Invalid tenant data provided to setTenant');
this.currentTenant = null;
localStorage.removeItem('currentTenant');
return;
}
this.currentTenant = tenant;
this._persistState('currentTenant', tenant);
},
loadTenant() {
const saved = this._loadState('currentTenant');
if (saved && typeof saved === 'object') {
this.currentTenant = saved;
} else {
this.currentTenant = null;
}
},
_cleanup() {
if (this._resizeHandler) {
window.removeEventListener('resize', this._resizeHandler);
}
if (this._keydownHandler) {
document.removeEventListener('keydown', this._keydownHandler);
}
if (this._htmxBeforeHandler) {
document.removeEventListener('htmx:beforeRequest', this._htmxBeforeHandler);
}
if (this._htmxAfterHandler) {
document.removeEventListener('htmx:afterRequest', this._htmxAfterHandler);
}
},
init() {
this.loadTheme();
this.loadTenant();
let resizeTimeout;
this._resizeHandler = () => {
clearTimeout(resizeTimeout);
resizeTimeout = setTimeout(() => {
if (window.innerWidth >= 768) {
this.closeMobileMenu();
}
}, 150);
};
window.addEventListener('resize', this._resizeHandler);
this._keydownHandler = (e) => {
if (e.key === 'Escape') {
if (this.activeModal) {
this.closeModal();
} else if (this.showNotifications) {
this.closeNotifications();
} else if (this.mobileMenuOpen) {
this.closeMobileMenu();
}
}
};
document.addEventListener('keydown', this._keydownHandler);
this._htmxBeforeHandler = () => {
this._pendingRequests++;
this.setLoading(true);
};
this._htmxAfterHandler = () => {
this._pendingRequests = Math.max(0, this._pendingRequests - 1);
if (this._pendingRequests === 0) {
this.setLoading(false);
}
};
document.addEventListener('htmx:beforeRequest', this._htmxBeforeHandler);
document.addEventListener('htmx:afterRequest', this._htmxAfterHandler);
document.addEventListener('htmx:responseError', this._htmxAfterHandler);
document.addEventListener('htmx:sendError', this._htmxAfterHandler);
window.addEventListener('beforeunload', () => {
this._cleanup();
});
}
});
});
// User Store
document.addEventListener('alpine:init', () => {
Alpine.store('user', {
isAuthenticated: false,
profile: null,
permissions: [],
roles: [],
preferences: {},
loginAttempts: 0,
maxLoginAttempts: 3,
lockoutTime: null,
sessionExpiry: null,
lastActivity: Date.now(),
_activityTimer: null,
_sessionCheckInterval: null,
_activityHandlers: [],
_htmxErrorHandler: null,
_lastActivityUpdate: 0,
_activityThrottle: 5000, // Update activity max once per 5 seconds
_persistState(key, value) {
try {
localStorage.setItem(key, typeof value === 'string' ? value : JSON.stringify(value));
return true;
} catch (e) {
console.error(`Failed to persist ${key}:`, e);
return false;
}
},
_loadState(key, defaultValue = null) {
try {
const saved = localStorage.getItem(key);
if (!saved) return defaultValue;
try {
return JSON.parse(saved);
} catch {
return saved;
}
} catch (e) {
console.error(`Failed to load ${key}:`, e);
return defaultValue;
}
},
_removeState(key) {
try {
localStorage.removeItem(key);
} catch (e) {
console.error(`Failed to remove ${key}:`, e);
}
},
_isValidUserData(userData) {
return userData &&
typeof userData === 'object' &&
userData.profile &&
typeof userData.profile === 'object';
},
_sanitizeArray(value) {
return Array.isArray(value) ? value.filter(item =>
item !== null && item !== undefined
) : [];
},
setUser(userData) {
if (!userData || typeof userData !== 'object') {
console.warn('Invalid user data provided to setUser');
return false;
}
this.profile = userData;
this.isAuthenticated = true;
this.loginAttempts = 0;
this.lockoutTime = null;
this.updateLastActivity();
this.persistUserData();
return true;
},
setPermissions(permissions) {
this.permissions = this._sanitizeArray(permissions);
this.persistUserData();
},
setRoles(roles) {
this.roles = this._sanitizeArray(roles);
this.persistUserData();
},
hasPermission(permission) {
if (!permission) return false;
return this.permissions.includes(permission);
},
hasRole(role) {
if (!role) return false;
return this.roles.includes(role);
},
hasAnyRole(roles) {
if (!Array.isArray(roles) || roles.length === 0) return false;
return roles.some(role => this.hasRole(role));
},
hasAllRoles(roles) {
if (!Array.isArray(roles) || roles.length === 0) return false;
return roles.every(role => this.hasRole(role));
},
updateProfile(profileData) {
if (!this.profile) {
console.warn('Cannot update profile: user not authenticated');
return false;
}
if (!profileData || typeof profileData !== 'object') {
console.warn('Invalid profile data provided');
return false;
}
this.profile = { ...this.profile, ...profileData };
this.persistUserData();
return true;
},
updatePreferences(prefs) {
if (!prefs || typeof prefs !== 'object') {
console.warn('Invalid preferences provided');
return false;
}
this.preferences = { ...this.preferences, ...prefs };
this.persistUserData();
return true;
},
getPreference(key, defaultValue = null) {
if (!key) return defaultValue;
return this.preferences.hasOwnProperty(key) ? this.preferences[key] : defaultValue;
},
incrementLoginAttempts() {
this.loginAttempts++;
if (this.loginAttempts >= this.maxLoginAttempts) {
this.lockoutTime = Date.now() + (15 * 60 * 1000); // 15 minutes
this._persistState('lockoutTime', this.lockoutTime);
this._persistState('loginAttempts', this.loginAttempts);
}
},
isLockedOut() {
if (!this.lockoutTime) return false;
const lockedOut = Date.now() < this.lockoutTime;
if (!lockedOut && this.lockoutTime) {
this.lockoutTime = null;
this.loginAttempts = 0;
this._removeState('lockoutTime');
this._removeState('loginAttempts');
}
return lockedOut;
},
getRemainingLockoutTime() {
if (!this.isLockedOut()) return 0;
return Math.ceil((this.lockoutTime - Date.now()) / 1000);
},
updateLastActivity() {
const now = Date.now();
if (now - this._lastActivityUpdate < this._activityThrottle) {
return;
}
this.lastActivity = now;
this._lastActivityUpdate = now;
this._persistState('userLastActivity', this.lastActivity.toString());
},
checkSessionExpiry() {
if (!this.isAuthenticated) return false;
if (this.sessionExpiry && Date.now() > this.sessionExpiry) {
console.info('Session expired');
this.logout();
return true;
}
return false;
},
extendSession(expiryTime) {
if (!expiryTime || typeof expiryTime !== 'number') {
console.warn('Invalid expiry time provided');
return false;
}
this.sessionExpiry = expiryTime;
this._persistState('sessionExpiry', expiryTime.toString());
return true;
},
logout() {
this.isAuthenticated = false;
this.profile = null;
this.permissions = [];
this.roles = [];
this.preferences = {};
this.sessionExpiry = null;
this.lastActivity = Date.now();
this.clearUserData();
window.dispatchEvent(new CustomEvent('user:logout'));
},
persistUserData() {
if (!this.isAuthenticated || !this.profile) {
return false;
}
const userData = {
profile: this.profile,
permissions: this.permissions,
roles: this.roles,
preferences: this.preferences,
lastActivity: this.lastActivity
};
return this._persistState('userData', userData);
},
loadUserData() {
const saved = this._loadState('userData');
if (saved && this._isValidUserData(saved)) {
this.profile = saved.profile;
this.permissions = this._sanitizeArray(saved.permissions);
this.roles = this._sanitizeArray(saved.roles);
this.preferences = saved.preferences || {};
this.lastActivity = saved.lastActivity || Date.now();
this.isAuthenticated = true;
}
const sessionExpiry = this._loadState('sessionExpiry');
if (sessionExpiry) {
this.sessionExpiry = parseInt(sessionExpiry);
}
const lockoutTime = this._loadState('lockoutTime');
if (lockoutTime) {
this.lockoutTime = parseInt(lockoutTime);
this.loginAttempts = parseInt(this._loadState('loginAttempts', 0));
}
},
clearUserData() {
this._removeState('userData');
this._removeState('sessionExpiry');
this._removeState('userLastActivity');
this._removeState('lockoutTime');
this._removeState('loginAttempts');
},
trackActivity() {
if (this.isAuthenticated) {
this.updateLastActivity();
}
},
_cleanup() {
if (this._activityTimer) {
clearTimeout(this._activityTimer);
this._activityTimer = null;
}
if (this._sessionCheckInterval) {
clearInterval(this._sessionCheckInterval);
this._sessionCheckInterval = null;
}
this._activityHandlers.forEach(({ event, handler }) => {
document.removeEventListener(event, handler);
});
this._activityHandlers = [];
if (this._htmxErrorHandler) {
document.removeEventListener('htmx:responseError', this._htmxErrorHandler);
this._htmxErrorHandler = null;
}
},
init() {
this.loadUserData();
this.checkSessionExpiry();
const activityEvents = ['click', 'keypress', 'scroll', 'mousemove'];
activityEvents.forEach(event => {
const handler = () => {
if (this._activityTimer) {
clearTimeout(this._activityTimer);
}
this._activityTimer = setTimeout(() => {
this.trackActivity();
}, 1000);
};
document.addEventListener(event, handler, { passive: true });
this._activityHandlers.push({ event, handler });
});
this._sessionCheckInterval = setInterval(() => {
this.checkSessionExpiry();
}, 60000); // Check every minute
this._htmxErrorHandler = (event) => {
if (event.detail.xhr && event.detail.xhr.status === 401) {
this.logout();
const shouldRedirect = window.dispatchEvent(
new CustomEvent('user:unauthorized', {
cancelable: true,
detail: { event }
})
);
if (shouldRedirect) {
window.location.href = '/auth/login';
}
}
};
document.addEventListener('htmx:responseError', this._htmxErrorHandler);
window.addEventListener('beforeunload', () => {
this._cleanup();
});
window.dispatchEvent(new CustomEvent('user:initialized', {
detail: { isAuthenticated: this.isAuthenticated }
}));
}
});
});
// Notifications Store
document.addEventListener('alpine:init', () => {
Alpine.store('notifications', {
items: [],
unreadCount: 0,
maxNotifications: 50,
isOpen: false,
filter: 'all', // 'all', 'unread', 'read'
sortBy: 'newest', // 'newest', 'oldest', 'priority'
autoDismissTimeout: 5000,
pauseOnHover: true,
types: {
success: { icon: 'check-circle', color: 'green', autoDismiss: true },
error: { icon: 'x-circle', color: 'red', autoDismiss: false },
warning: { icon: 'exclamation-triangle', color: 'yellow', autoDismiss: false },
info: { icon: 'information-circle', color: 'blue', autoDismiss: true }
},
_timers: new Map(), // Track auto-dismiss timers
_cleanupInterval: null,
_htmxAfterHandler: null,
_htmxErrorHandler: null,
_persistenceTimer: null,
_persistenceDebounce: 300, // Debounce persistence writes
_cachedFiltered: null,
_cacheInvalidated: true,
_persistState(key, value) {
try {
const serialized = typeof value === 'string' ? value : JSON.stringify(value);
localStorage.setItem(key, serialized);
return true;
} catch (e) {
console.error(`Failed to persist ${key}:`, e);
if (e.name === 'QuotaExceededError') {
this._handleStorageQuotaExceeded();
}
return false;
}
},
_loadState(key, defaultValue = null) {
try {
const saved = localStorage.getItem(key);
if (!saved) return defaultValue;
try {
return JSON.parse(saved);
} catch {
return saved;
}
} catch (e) {
console.error(`Failed to load ${key}:`, e);
return defaultValue;
}
},
_removeState(key) {
try {
localStorage.removeItem(key);
} catch (e) {
console.error(`Failed to remove ${key}:`, e);
}
},
_handleStorageQuotaExceeded() {
const readNotifications = this.items.filter(n => n.read);
if (readNotifications.length > 0) {
const toRemove = Math.ceil(readNotifications.length / 2);
for (let i = 0; i < toRemove; i++) {
const oldest = readNotifications[readNotifications.length - 1 - i];
this.dismiss(oldest.id, { skipPersist: true });
}
this._persistNotifications({ immediate: true });
}
},
_isValidNotification(notification) {
return notification &&
typeof notification === 'object' &&
notification.message &&
typeof notification.message === 'string' &&
notification.message.trim().length > 0;
},
_sanitizeNotification(notification) {
const sanitized = { ...notification };
if (Array.isArray(sanitized.actions)) {
sanitized.actions = sanitized.actions.map(action => {
const { callback, ...rest } = action;
return rest; // Remove callback function
});
}
return sanitized;
},
_invalidateCache() {
this._cacheInvalidated = true;
this._cachedFiltered = null;
},
add(notification) {
if (!this._isValidNotification(notification)) {
console.warn('Invalid notification data');
return null;
}
const id = this._generateId();
const now = Date.now();
const type = notification.type || 'info';
const config = this.types[type] || this.types.info;
const newNotification = {
id,
message: notification.message.trim(),
title: notification.title || '',
type,
icon: notification.icon || config.icon,
color: notification.color || config.color,
read: false,
persistent: notification.persistent || false,
priority: notification.priority || 'normal',
timestamp: now,
expiresAt: notification.expiresAt || null,
actions: notification.actions || [],
data: notification.data || {},
autoDismiss: notification.hasOwnProperty('autoDismiss')
? notification.autoDismiss
: config.autoDismiss
};
this.items.unshift(newNotification);
this.unreadCount++;
this._invalidateCache();
if (this.items.length > this.maxNotifications) {
const removed = this.items.splice(this.maxNotifications);
removed.forEach(item => {
if (!item.read) {
this.unreadCount = Math.max(0, this.unreadCount - 1);
}
this._cancelTimer(item.id);
});
}
this._persistNotifications();
if (newNotification.autoDismiss && !newNotification.persistent) {
const timer = setTimeout(() => {
this.dismiss(id);
}, this.autoDismissTimeout);
this._timers.set(id, timer);
}
window.dispatchEvent(new CustomEvent('notification:added', {
detail: { id, notification: newNotification }
}));
return id;
},
success(message, options = {}) {
return this.add({ ...options, message, type: 'success' });
},
error(message, options = {}) {
return this.add({ ...options, message, type: 'error' });
},
warning(message, options = {}) {
return this.add({ ...options, message, type: 'warning' });
},
info(message, options = {}) {
return this.add({ ...options, message, type: 'info' });
},
dismiss(id, options = {}) {
const index = this.items.findIndex(item => item.id === id);
if (index === -1) return false;
const notification = this.items[index];
this._cancelTimer(id);
if (!notification.read) {
this.unreadCount = Math.max(0, this.unreadCount - 1);
}
this.items.splice(index, 1);
this._invalidateCache();
if (!options.skipPersist) {
this._persistNotifications();
}
window.dispatchEvent(new CustomEvent('notification:dismissed', {
detail: { id, notification }
}));
return true;
},
markAsRead(id) {
const notification = this.items.find(item => item.id === id);
if (!notification || notification.read) return false;
notification.read = true;
this.unreadCount = Math.max(0, this.unreadCount - 1);
this._invalidateCache();
this._persistNotifications();
window.dispatchEvent(new CustomEvent('notification:read', {
detail: { id, notification }
}));
return true;
},
markAllAsRead() {
let marked = 0;
this.items.forEach(item => {
if (!item.read) {
item.read = true;
marked++;
}
});
if (marked > 0) {
this.unreadCount = 0;
this._invalidateCache();
this._persistNotifications();
window.dispatchEvent(new CustomEvent('notification:allRead', {
detail: { count: marked }
}));
}
return marked;
},
clear() {
this._timers.forEach((timer, id) => {
clearTimeout(timer);
});
this._timers.clear();
const count = this.items.length;
this.items = [];
this.unreadCount = 0;
this._invalidateCache();
this._persistNotifications();
window.dispatchEvent(new CustomEvent('notification:cleared', {
detail: { count }
}));
},
clearRead() {
const beforeCount = this.items.length;
const readIds = this.items.filter(item => item.read).map(item => item.id);
readIds.forEach(id => this._cancelTimer(id));
this.items = this.items.filter(item => !item.read);
this._invalidateCache();
if (beforeCount !== this.items.length) {
this._persistNotifications();
window.dispatchEvent(new CustomEvent('notification:readCleared', {
detail: { count: readIds.length }
}));
}
},
get filteredItems() {
if (!this._cacheInvalidated && this._cachedFiltered) {
return this._cachedFiltered;
}
let filtered = [...this.items];
switch (this.filter) {
case 'unread':
filtered = filtered.filter(item => !item.read);
break;
case 'read':
filtered = filtered.filter(item => item.read);
break;
}
switch (this.sortBy) {
case 'oldest':
filtered.sort((a, b) => a.timestamp - b.timestamp);
break;
case 'priority':
const priorityOrder = { high: 3, normal: 2, low: 1 };
filtered.sort((a, b) => {
const aPriority = priorityOrder[a.priority] || 2;
const bPriority = priorityOrder[b.priority] || 2;
if (aPriority === bPriority) {
return b.timestamp - a.timestamp;
}
return bPriority - aPriority;
});
break;
default: // 'newest'
filtered.sort((a, b) => b.timestamp - a.timestamp);
}
this._cachedFiltered = filtered;
this._cacheInvalidated = false;
return filtered;
},
get hasUnread() {
return this.unreadCount > 0;
},
get isEmpty() {
return this.items.length === 0;
},
toggle() {
this.isOpen = !this.isOpen;
},
open() {
this.isOpen = true;
},
close() {
this.isOpen = false;
},
setFilter(filter) {
if (['all', 'unread', 'read'].includes(filter) && this.filter !== filter) {
this.filter = filter;
this._invalidateCache();
}
},
setSortBy(sortBy) {
if (['newest', 'oldest', 'priority'].includes(sortBy) && this.sortBy !== sortBy) {
this.sortBy = sortBy;
this._invalidateCache();
}
},
executeAction(notificationId, actionId) {
const notification = this.items.find(item => item.id === notificationId);
if (!notification) return false;
const action = notification.actions.find(a => a.id === actionId);
if (!action) return false;
this.markAsRead(notificationId);
if (typeof action.callback === 'function') {
try {
action.callback(notification, action);
} catch (e) {
console.error('Error executing notification action:', e);
return false;
}
}
if (action.dismiss) {
this.dismiss(notificationId);
}
window.dispatchEvent(new CustomEvent('notification:actionExecuted', {
detail: { notificationId, actionId, action }
}));
return true;
},
_cancelTimer(id) {
const timer = this._timers.get(id);
if (timer) {
clearTimeout(timer);
this._timers.delete(id);
}
},
pauseAutoDismiss(id) {
this._cancelTimer(id);
},
resumeAutoDismiss(id) {
const notification = this.items.find(item => item.id === id);
if (notification && notification.autoDismiss && !notification.persistent) {
const timer = setTimeout(() => {
this.dismiss(id);
}, this.autoDismissTimeout);
this._timers.set(id, timer);
}
},
_generateId() {
return `notif_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
},
_persistNotifications(options = {}) {
if (this._persistenceTimer) {
clearTimeout(this._persistenceTimer);
}
const doPersist = () => {
const data = {
items: this.items.map(item => this._sanitizeNotification(item)),
unreadCount: this.unreadCount
};
this._persistState('notifications', data);
};
if (options.immediate) {
doPersist();
} else {
this._persistenceTimer = setTimeout(doPersist, this._persistenceDebounce);
}
},
_loadNotifications() {
const data = this._loadState('notifications');
if (data && typeof data === 'object') {
this.items = Array.isArray(data.items) ? data.items : [];
this.unreadCount = typeof data.unreadCount === 'number' ? data.unreadCount : 0;
this.items = this.items.filter(item => {
return item &&
item.id &&
item.message &&
typeof item.timestamp === 'number';
});
const actualUnread = this.items.filter(item => !item.read).length;
if (actualUnread !== this.unreadCount) {
this.unreadCount = actualUnread;
}
this._cleanupExpired();
this._invalidateCache();
}
},
_cleanupExpired() {
const now = Date.now();
const beforeCount = this.items.length;
this.items = this.items.filter(item => {
const expired = item.expiresAt && now > item.expiresAt;
if (expired) {
if (!item.read) {
this.unreadCount = Math.max(0, this.unreadCount - 1);
}
this._cancelTimer(item.id);
return false;
}
return true;
});
if (this.items.length !== beforeCount) {
this._invalidateCache();
this._persistNotifications({ immediate: true });
}
},
_cleanup() {
this._timers.forEach(timer => clearTimeout(timer));
this._timers.clear();
if (this._cleanupInterval) {
clearInterval(this._cleanupInterval);
this._cleanupInterval = null;
}
if (this._persistenceTimer) {
clearTimeout(this._persistenceTimer);
this._persistenceTimer = null;
}
if (this._htmxAfterHandler) {
document.removeEventListener('htmx:afterRequest', this._htmxAfterHandler);
this._htmxAfterHandler = null;
}
if (this._htmxErrorHandler) {
document.removeEventListener('htmx:responseError', this._htmxErrorHandler);
this._htmxErrorHandler = null;
}
},
syncWithServer() {
console.info('Server sync not yet implemented');
},
markAsReadOnServer(id) {
console.info('Server-side read status not yet implemented');
},
init() {
this._loadNotifications();
this._cleanupInterval = setInterval(() => {
this._cleanupExpired();
}, 60000); // Check every minute
this._htmxAfterHandler = (event) => {
const response = event.detail.xhr;
if (!response) return;
const notificationHeader = response.getResponseHeader('X-Notification');
if (notificationHeader) {
try {
const notification = JSON.parse(notificationHeader);
this.add(notification);
} catch (e) {
console.error('Failed to parse server notification:', e);
}
}
};
document.addEventListener('htmx:afterRequest', this._htmxAfterHandler);
this._htmxErrorHandler = (event) => {
const xhr = event.detail.xhr;
if (!xhr) return;
const status = xhr.status;
let message = 'An error occurred';
switch (status) {
case 400:
message = 'Bad request. Please check your input.';
break;
case 401:
message = 'Authentication required. Please log in.';
break;
case 403:
message = 'You do not have permission to perform this action.';
break;
case 404:
message = 'The requested resource was not found.';
break;
case 500:
message = 'Server error. Please try again later.';
break;
default:
if (status >= 500) {
message = `Server error ${status}. Please try again later.`;
} else if (status >= 400) {
message = `Request failed with error ${status}.`;
}
}
this.error(message, {
title: 'Request Failed',
persistent: status >= 500,
data: {
status,
url: event.detail.requestConfig?.path || event.detail.pathInfo?.requestPath
}
});
};
document.addEventListener('htmx:responseError', this._htmxErrorHandler);
window.addEventListener('beforeunload', () => {
this._cleanup();
});
window.dispatchEvent(new CustomEvent('notifications:initialized', {
detail: {
count: this.items.length,
unread: this.unreadCount
}
}));
}
});
});
// === ESSENTIAL UTILITIES ===
// Form Utilities
function tenantForm() {
return {
formValid: false,
initForm() {
this.validateForm();
this.$el.addEventListener('input', () => {
this.validateForm();
});
const nameField = this.$el.querySelector('[name="name"]');
const subdomainField = this.$el.querySelector('[name="subdomain"]');
if (nameField && subdomainField && !subdomainField.value) {
nameField.addEventListener('input', (e) => {
if (!subdomainField.value) {
const subdomain = this.generateSubdomain(e.target.value);
subdomainField.value = subdomain;
subdomainField.dispatchEvent(new Event('input', { bubbles: true }));
}
});
}
},
validateForm() {
const form = this.$el.closest('form');
if (!form) return;
const requiredFields = form.querySelectorAll('[required]');
let allValid = true;
requiredFields.forEach(field => {
if (!field.value.trim()) {
allValid = false;
}
});
this.formValid = allValid;
const submitButton = form.querySelector('[type="submit"]');
if (submitButton) {
submitButton.disabled = !allValid;
submitButton.classList.toggle('opacity-50', !allValid);
submitButton.classList.toggle('cursor-not-allowed', !allValid);
}
},
generateSubdomain(name) {
return name
.toLowerCase()
.replace(/[^a-z0-9\s-]/g, '') // Remove special characters
.replace(/\s+/g, '-') // Replace spaces with hyphens
.replace(/-+/g, '-') // Replace multiple hyphens with single
.replace(/^-|-$/g, '') // Remove leading/trailing hyphens
.substring(0, 50); // Limit length
}
}
}
function userForm() {
return {
formValid: false,
initForm() {
this.validateForm();
this.$el.addEventListener('input', () => {
this.validateForm();
});
const emailField = this.$el.querySelector('[name="email"]');
const usernameField = this.$el.querySelector('[name="username"]');
if (emailField && usernameField && !usernameField.value) {
emailField.addEventListener('input', (e) => {
if (!usernameField.value) {
const username = this.generateUsername(e.target.value);
usernameField.value = username;
usernameField.dispatchEvent(new Event('input', { bubbles: true }));
}
});
}
},
validateForm() {
const form = this.$el.querySelector('form');
if (!form) return;
const requiredFields = form.querySelectorAll('[required]');
let allValid = true;
requiredFields.forEach(field => {
if (!field.value.trim()) {
allValid = false;
}
});
const password = form.querySelector('[name="password"]');
const passwordConfirm = form.querySelector('[name="password_confirm"]');
if (password && passwordConfirm && password.value !== passwordConfirm.value) {
allValid = false;
}
this.formValid = allValid;
},
generateUsername(email) {
const username = email.split('@')[0];
return username
.toLowerCase()
.replace(/[^a-z0-9]/g, '')
.substring(0, 20);
},
handleSubmit(e) {
if (!this.formValid) {
e.preventDefault();
Alpine.store('app').addNotification('error', 'Please fix form errors before submitting');
}
}
}
}
function passwordStrength() {
return {
password: '',
strength: 'weak',
strengthPercent: 0,
hasMinLength: false,
hasUppercase: false,
hasLowercase: false,
hasNumber: false,
hasSpecial: false,
init() {
this.$watch('password', () => {
this.calculateStrength();
});
const passwordField = this.$el.querySelector('[name="password"]');
if (passwordField) {
passwordField.addEventListener('input', (e) => {
this.password = e.target.value;
});
}
},
calculateStrength() {
this.hasMinLength = this.password.length >= 8;
this.hasUppercase = /[A-Z]/.test(this.password);
this.hasLowercase = /[a-z]/.test(this.password);
this.hasNumber = /[0-9]/.test(this.password);
this.hasSpecial = /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(this.password);
const score = [
this.hasMinLength,
this.hasUppercase,
this.hasLowercase,
this.hasNumber,
this.hasSpecial
].filter(Boolean).length;
if (score < 3) {
this.strength = 'weak';
this.strengthPercent = 25;
} else if (score < 5) {
this.strength = 'medium';
this.strengthPercent = 60;
} else {
this.strength = 'strong';
this.strengthPercent = 100;
}
}
}
}
window.formUtils = {
enableAutoSave(formId, interval = 30000) {
const form = document.getElementById(formId);
if (!form) return;
setInterval(() => {
const formData = new FormData(form);
const data = {};
for (let [key, value] of formData.entries()) {
data[key] = value;
}
Alpine.store('forms').saveDraft(formId, data);
}, interval);
},
restoreFromDraft(formId) {
const draft = Alpine.store('forms').loadDraft(formId);
if (!draft) return false;
const form = document.getElementById(formId);
if (!form) return false;
Object.keys(draft).forEach(key => {
const field = form.querySelector(`[name="${key}"]`);
if (field) {
field.value = draft[key];
field.dispatchEvent(new Event('input', { bubbles: true }));
}
});
return true;
},
validateForm(formId) {
const form = document.getElementById(formId);
if (!form) return false;
const requiredFields = form.querySelectorAll('[required]');
let isValid = true;
requiredFields.forEach(field => {
if (!field.value.trim()) {
isValid = false;
field.classList.add('border-red-500');
} else {
field.classList.remove('border-red-500');
}
});
return isValid;
},
showFormErrors(errors) {
Object.keys(errors).forEach(fieldName => {
const field = document.querySelector(`[name="${fieldName}"]`);
if (field) {
const errorDiv = field.parentNode.querySelector('.text-red-500');
if (errorDiv) {
errorDiv.textContent = errors[fieldName];
errorDiv.style.display = 'block';
}
}
});
},
clearFormErrors(formId) {
const form = document.getElementById(formId);
if (!form) return;
const errorDivs = form.querySelectorAll('.text-red-500');
errorDivs.forEach(div => {
div.textContent = '';
div.style.display = 'none';
});
const errorFields = form.querySelectorAll('.border-red-500');
errorFields.forEach(field => {
field.classList.remove('border-red-500');
});
}
};
// DataTable Utilities
function dataTable(config) {
return {
loading: false,
searchTerm: '',
selectedRows: [],
allSelected: false,
activeFilter: 'all',
sortColumn: '',
sortDirection: 'asc',
rows: [],
searchUrl: config.searchUrl || '',
filterUrl: config.filterUrl || '',
sortUrl: config.sortUrl || '',
initTable() {
this.loadInitialData();
this.setupKeyboardShortcuts();
},
loadInitialData() {
},
performSearch() {
if (!this.searchUrl) {
this.filterRowsLocally();
return;
}
this.loading = true;
htmx.ajax('GET', this.searchUrl, {
values: {
search: this.searchTerm,
filter: this.activeFilter,
sort: this.sortColumn,
direction: this.sortDirection
},
target: '#table-body',
swap: 'innerHTML'
}).then(() => {
this.loading = false;
}).catch(() => {
this.loading = false;
Alpine.store('app').addNotification('error', 'Failed to search data');
});
},
filterRowsLocally() {
},
applyFilter() {
if (!this.filterUrl) {
this.filterRowsLocally();
return;
}
this.loading = true;
htmx.ajax('GET', this.filterUrl, {
values: {
filter: this.activeFilter,
search: this.searchTerm,
sort: this.sortColumn,
direction: this.sortDirection
},
target: '#table-body',
swap: 'innerHTML'
}).then(() => {
this.loading = false;
}).catch(() => {
this.loading = false;
Alpine.store('app').addNotification('error', 'Failed to filter data');
});
},
getActiveFilterLabel() {
const filters = {
'all': 'All Items',
'active': 'Active',
'inactive': 'Inactive',
'recent': 'Recent',
'archived': 'Archived'
};
return filters[this.activeFilter] || 'Filter';
},
toggleSort(column) {
if (this.sortColumn === column) {
this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
} else {
this.sortColumn = column;
this.sortDirection = 'asc';
}
if (!this.sortUrl) {
this.sortRowsLocally();
return;
}
this.loading = true;
htmx.ajax('GET', this.sortUrl, {
values: {
sort: this.sortColumn,
direction: this.sortDirection,
search: this.searchTerm,
filter: this.activeFilter
},
target: '#table-body',
swap: 'innerHTML'
}).then(() => {
this.loading = false;
}).catch(() => {
this.loading = false;
Alpine.store('app').addNotification('error', 'Failed to sort data');
});
},
sortRowsLocally() {
},
toggleAll() {
if (this.allSelected) {
const checkboxes = document.querySelectorAll('input[type="checkbox"][id^="checkbox-table-"]');
checkboxes.forEach(checkbox => {
checkbox.checked = true;
const rowId = this.getRowIdFromCheckbox(checkbox);
if (rowId && !this.selectedRows.includes(rowId)) {
this.selectedRows.push(rowId);
}
});
} else {
this.selectedRows = [];
const checkboxes = document.querySelectorAll('input[type="checkbox"][id^="checkbox-table-"]');
checkboxes.forEach(checkbox => {
checkbox.checked = false;
});
}
},
updateSelectedRows(rowId, selected) {
if (selected) {
if (!this.selectedRows.includes(rowId)) {
this.selectedRows.push(rowId);
}
} else {
this.selectedRows = this.selectedRows.filter(id => id !== rowId);
this.allSelected = false;
}
const totalCheckboxes = document.querySelectorAll('input[type="checkbox"][id^="checkbox-table-"]').length;
this.allSelected = this.selectedRows.length === totalCheckboxes && totalCheckboxes > 0;
},
getRowIdFromCheckbox(checkbox) {
const row = checkbox.closest('tr');
return row ? row.getAttribute('data-row-id') : null;
},
executeBulkAction(action) {
if (this.selectedRows.length === 0) {
Alpine.store('app').addNotification('warning', 'Please select items first');
return;
}
switch (action) {
case 'delete':
this.confirmBulkDelete();
break;
case 'export':
this.exportSelected();
break;
case 'archive':
this.archiveSelected();
break;
default:
this.executeCustomBulkAction(action);
}
},
confirmBulkDelete() {
if (confirm(`Are you sure you want to delete ${this.selectedRows.length} items?`)) {
this.performBulkAction('delete');
}
},
exportSelected() {
const params = new URLSearchParams();
this.selectedRows.forEach(id => params.append('ids[]', id));
const downloadUrl = `/export?${params.toString()}`;
window.open(downloadUrl, '_blank');
},
archiveSelected() {
this.performBulkAction('archive');
},
executeCustomBulkAction(action) {
this.performBulkAction(action);
},
performBulkAction(action) {
this.loading = true;
htmx.ajax('POST', `/bulk-actions/${action}`, {
values: {
ids: this.selectedRows,
action: action
},
headers: {
'X-CSRF-Token': document.querySelector('meta[name=csrf-token]').content
},
target: '#table-body',
swap: 'innerHTML'
}).then(() => {
this.loading = false;
this.selectedRows = [];
this.allSelected = false;
Alpine.store('app').addNotification('success', `Bulk ${action} completed successfully`);
}).catch(() => {
this.loading = false;
Alpine.store('app').addNotification('error', `Failed to perform bulk ${action}`);
});
},
executeRowAction(action, rowId) {
switch (action) {
case 'delete':
this.confirmRowDelete(rowId);
break;
case 'duplicate':
this.duplicateRow(rowId);
break;
default:
this.performRowAction(action, rowId);
}
},
confirmRowDelete(rowId) {
if (confirm('Are you sure you want to delete this item?')) {
this.performRowAction('delete', rowId);
}
},
duplicateRow(rowId) {
this.performRowAction('duplicate', rowId);
},
performRowAction(action, rowId) {
this.loading = true;
htmx.ajax('POST', `/row-actions/${action}`, {
values: {
id: rowId,
action: action
},
headers: {
'X-CSRF-Token': document.querySelector('meta[name=csrf-token]').content
},
target: '#table-body',
swap: 'innerHTML'
}).then(() => {
this.loading = false;
Alpine.store('app').addNotification('success', `${action} completed successfully`);
}).catch(() => {
this.loading = false;
Alpine.store('app').addNotification('error', `Failed to perform ${action}`);
});
},
openActionModal(modalId, rowId) {
htmx.ajax('GET', `/modals/${modalId}`, {
values: { id: rowId },
target: '#modal-container',
swap: 'innerHTML'
});
},
goToPage(page) {
this.loading = true;
const currentUrl = new URL(window.location);
currentUrl.searchParams.set('page', page);
htmx.ajax('GET', currentUrl.toString(), {
target: '#table-body',
swap: 'innerHTML'
}).then(() => {
this.loading = false;
window.history.pushState({}, '', currentUrl.toString());
}).catch(() => {
this.loading = false;
Alpine.store('app').addNotification('error', 'Failed to load page');
});
},
setupKeyboardShortcuts() {
document.addEventListener('keydown', (e) => {
if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA') {
return;
}
switch (e.key) {
case 'f':
if (e.ctrlKey || e.metaKey) {
e.preventDefault();
document.getElementById('table-search')?.focus();
}
break;
case 'a':
if (e.ctrlKey || e.metaKey) {
e.preventDefault();
this.allSelected = !this.allSelected;
this.toggleAll();
}
break;
case 'Escape':
this.selectedRows = [];
this.allSelected = false;
document.querySelectorAll('input[type="checkbox"]').forEach(cb => cb.checked = false);
break;
}
});
}
}
}
window.refreshTable = function() {
window.location.reload();
};
window.clearTableFilters = function() {
const tableComponent = document.querySelector('[x-data]').__x?.$data;
if (tableComponent) {
tableComponent.searchTerm = '';
tableComponent.activeFilter = 'all';
tableComponent.sortColumn = '';
tableComponent.sortDirection = 'asc';
tableComponent.performSearch();
}
};