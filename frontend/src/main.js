import './style.css';
import './app.css';

import logo from './assets/images/logo-universal.png';
import {
    GetSystemInfo,
    ScanConversations,
    RunFix,
    CheckForUpdates,
    AssignWorkspace,
    SelectFolder
} from '../wailsjs/go/main/App.js';
import { EventsOn } from '../wailsjs/runtime/runtime.js';

// Translation dictionaries
const TRANSLATIONS = {
    en: {
        app_title: "GravityAnchor",
        step_system_check: "System Check",
        step_scan_resolve: "Scan & Resolve",
        step_rebuild_db: "Rebuild Database",
        welcome_title: "Fix Your Antigravity Conversations",
        welcome_desc: "Rebuilds your local Antigravity workspace index, maps missing workspaces, restores conversation titles from brain artifacts, and injects timestamps for correct list sorting.",
        system_paths_title: "System Paths Check",
        sys_checking: "Checking system configuration and path detection...",
        btn_scan_conversations: "Scan Conversations",
        stat_brain_titles: "Brain Titles",
        stat_preserved_titles: "Preserved Titles",
        stat_fallback_titles: "Fallback Titles",
        stat_mapped_workspaces: "Mapped Workspaces",
        btn_rescan: "Rescan",
        th_index: "Index",
        th_status: "Status",
        th_details: "Conversation Details",
        th_workspace: "Workspace Association",
        th_action: "Action",
        mode_auto: "Auto Workspace Mapping (Recommended)",
        mode_strict: "Strict Mode (Require Workspaces)",
        btn_rebuild_fix: "🚀 Rebuild & Fix Index",
        progress_title: "Rebuilding Trajectory Index...",
        progress_status: "Preparing databases and backups...",
        success_title: "Index Rebuilt Successfully!",
        success_desc: "Databases successfully patched.",
        detail_total: "Total Rebuilt",
        detail_ws: "Workspace Mappings",
        detail_ts: "Timestamps Injected",
        detail_status: "Status",
        detail_perfect: "✅ Perfect",
        btn_return_dashboard: "Perfect! Return to Dashboard",
        search_placeholder: "Search conversations by title or ID...",
        filter_all: "All Conversations",
        filter_missing_ws: "Missing Workspace Mapping",
        filter_has_ws: "Has Workspace Mapping",
        filter_source_brain: "Title from Brain",
        filter_source_preserved: "Title from DB",
        filter_source_fallback: "Title from Fallback",
        dialog_select_title: "Select Workspace Folder",
        btn_map_dir: "Map Dir",
        btn_reassign: "Reassign",
        badge_new_title: "New entry to build from Brain file",
        badge_exist_title: "Existing record to update",
        badge_fallback_title: "No title metadata. Uses Fallback.",
        btn_back_dashboard: "Back to Dashboard",
        fix_failed: "Fix Execution Failed",
        fix_failed_desc: "An error occurred during build processes.",
        os_badge_prefix: "OS: ",
        status_active: "Workspace Active",
        status_inactive: "No Workspace",
        lbl_sqlite_found: "SQLite state.vscdb found in {count} location(s).",
        lbl_sqlite_missing: "No active state.vscdb database path detected. Antigravity might not be configured.",
        lbl_conv_found: "Detected {count} conversation search director(ies).",
        lbl_conv_missing: "No standard Antigravity conversation storage directories detected.",
        lbl_brain_found: "Detected active Brain workspace storage directory.",
        lbl_brain_missing: "No local Brain workspace directories detected. Auto title restoration may be limited.",
        lbl_scanning: "Scanning conversation data. Please wait...",
        lbl_scan_failed: "Scan Failed",
        lbl_no_conv: "No conversations match the current criteria.",
        log_initialized: "Starting engine transaction...",
        log_manual_assign: "Assigned manual workspace path for {cid}: {dirPath}",
        log_scanning_backend: "Scanning conversation directories...",
        log_found_conv: "Found {count} unique conversations",
        log_reading_meta: "Reading metadata from database(s)...",
        log_titles_stats: "Titles: {brain} brain, {preserved} preserved, {fallback} fallback",
        log_start_fix: "Starting fix process...",
        log_auto_mapping: "Auto-assigning workspaces for {count} conversations...",
        log_auto_mapped: "Auto-assigned {count} workspace(s)",
        log_building_index: "Building final index...",
        log_processed: "Processing {step}/{total}...",
        log_workspace_summary: "Workspace: {mapped} mapped | Timestamps injected: {injected}",
        log_writing_db: "Writing to database(s)...",
        log_updated_db: "Updated: {appName}",
        log_db_no_change: "No SQLite databases modified. (Already fully in sync!)",
        log_success: "SUCCESS! Rebuilt index with {count} conversations.",
        err_first_scan: "Please scan conversations first",
        err_no_db: "No database found"
    },
    zh: {
        app_title: "GravityAnchor 对话修复工具",
        step_system_check: "系统检查",
        step_scan_resolve: "扫描与解析",
        step_rebuild_db: "重建数据库",
        welcome_title: "使用 GravityAnchor 修复你的对话列表",
        welcome_desc: "一键重建本地 Antigravity 工作区索引，智能映射缺失的工作区，从 brain 文件中恢复对话标题，并注入修改时间戳以保证正确的列表排序。",
        system_paths_title: "系统路径检测",
        sys_checking: "正在检查系统配置和路径探测...",
        btn_scan_conversations: "开始扫描对话",
        stat_brain_titles: "Brain 标题",
        stat_preserved_titles: "DB 保留标题",
        stat_fallback_titles: "兜底标题",
        stat_mapped_workspaces: "已映射工作区",
        btn_rescan: "重新扫描",
        th_index: "序号",
        th_status: "状态",
        th_details: "对话详情",
        th_workspace: "关联工作区",
        th_action: "操作",
        mode_auto: "自动映射工作区 (推荐)",
        mode_strict: "严格模式 (必须关联工作区)",
        btn_rebuild_fix: "🚀 重建并修复索引",
        progress_title: "正在重建对话索引...",
        progress_status: "正在准备数据库和备份文件...",
        success_title: "对话索引重建成功！",
        success_desc: "已成功修补所有配置的数据库条目。",
        detail_total: "重建对话总数",
        detail_ws: "已关联工作区",
        detail_ts: "注入的时间戳",
        detail_status: "执行状态",
        detail_perfect: "✅ 完美修复",
        btn_return_dashboard: "太棒了！返回仪表盘",
        search_placeholder: "按标题或对话 ID 搜索对话...",
        filter_all: "全部对话列表",
        filter_missing_ws: "未映射工作区",
        filter_has_ws: "已映射工作区",
        filter_source_brain: "标题源自 Brain",
        filter_source_preserved: "标题源自 DB",
        filter_source_fallback: "标题源自兜底",
        dialog_select_title: "请选择该对话对应的工作区文件夹",
        btn_map_dir: "关联目录",
        btn_reassign: "重新关联",
        badge_new_title: "从 Brain 文件检测到的新对话",
        badge_exist_title: "已存在并待更新的对话",
        badge_fallback_title: "无有效标题，采用时间兜底",
        btn_back_dashboard: "返回仪表盘",
        fix_failed: "修复执行失败",
        fix_failed_desc: "在执行构建和数据库写入时发生错误。",
        os_badge_prefix: "系统: ",
        status_active: "已关联工作区",
        status_inactive: "未关联工作区",
        lbl_sqlite_found: "在 {count} 处检测到 SQLite 状态数据库 state.vscdb。",
        lbl_sqlite_missing: "未检测到活跃的 state.vscdb 路径，Antigravity 可能尚未配置。",
        lbl_conv_found: "检测到 {count} 处 Antigravity 对话文件搜索存储目录。",
        lbl_conv_missing: "未发现标准的 Antigravity 对话存储目录。",
        lbl_brain_found: "已检测到活跃的 Brain 工作区文档存储路径。",
        lbl_brain_missing: "未检测到本地 Brain 路径。自动解析标题可能受限。",
        lbl_scanning: "正在扫描对话数据，请稍候...",
        lbl_scan_failed: "扫描失败",
        lbl_no_conv: "没有找到符合当前过滤条件的对话。",
        log_initialized: "正在初始化数据库事务...",
        log_manual_assign: "已为对话 {cid} 手动分配工作区目录: {dirPath}",
        log_scanning_backend: "正在扫描对话存储目录...",
        log_found_conv: "共发现 {count} 个唯一的对话记录",
        log_reading_meta: "正在从数据库读取已保存的元数据...",
        log_titles_stats: "标题解析：{brain} 处源于 Brain，{preserved} 处保留，{fallback} 处兜底",
        log_start_fix: "开始执行对话修复流程...",
        log_auto_mapping: "正在为 {count} 个未映射的对话自动解析工作区...",
        log_auto_mapped: "已自动映射 {count} 个工作区",
        log_building_index: "正在构建最终对话索引...",
        log_processed: "正在处理 {step}/{total}...",
        log_workspace_summary: "工作区关联：已映射 {mapped} 个 | 注入修改时间戳：{injected} 个",
        log_writing_db: "正在将配置写入 SQLite 数据库...",
        log_updated_db: "已更新数据库索引：{appName}",
        log_db_no_change: "未更改任何数据库。（数据已处于最新同步状态！）",
        log_success: "成功！已重建具有 {count} 个对话的轨迹索引。",
        err_first_scan: "请先执行对话扫描",
        err_no_db: "未找到任何可用数据库"
    }
};

// Application State
let state = {
    conversations: [],
    filteredConversations: [],
    systemInfo: null,
    searchQuery: '',
    statusFilter: 'all',
    activeStep: 1,
    lang: 'en'
};

// SVG Icons as String Constants
const ICONS = {
    checkCircle: `<svg style="color: var(--accent-success);" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>`,
    alertCircle: `<svg style="color: var(--accent-warning);" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>`,
    folder: `<svg style="opacity: 0.65; margin-right: 4px;" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path></svg>`
};

// Initialize Application
window.addEventListener('DOMContentLoaded', () => {
    // Inject Logo Assets
    const logoEl = document.getElementById('logo');
    const logoHeroEl = document.getElementById('logo-hero');
    if (logoEl) logoEl.src = logo;
    if (logoHeroEl) logoHeroEl.src = logo;

    // Detect language
    detectLanguage();

    // Load initial system information
    loadSystemInfo();

    // Subscribe to Wails events
    setupEventSubscriptions();

    // Perform check for application updates
    runUpdateCheck();
});

// Detect user's language (LocalStorage -> Navigator Language -> Default English)
function detectLanguage() {
    let savedLang = localStorage.getItem('app-lang');
    if (!savedLang) {
        // Default based on system locale navigator language
        savedLang = navigator.language && navigator.language.startsWith('zh') ? 'zh' : 'en';
    }
    state.lang = savedLang;
    updateLanguageUI();
}

// Global change language trigger
window.changeLanguage = function(lang) {
    state.lang = lang;
    localStorage.setItem('app-lang', lang);
    updateLanguageUI();
    
    // Refresh current screens dynamically
    if (state.systemInfo) {
        loadSystemInfo();
    }
    if (state.conversations.length > 0) {
        filterConversations();
    }
};

// Dynamic translator that updates DOM elements marked with data-i18n
function updateLanguageUI() {
    const lang = state.lang;
    const t = TRANSLATIONS[lang];
    if (!t) return;

    // Translate standard static texts
    document.querySelectorAll('[data-i18n]').forEach((el) => {
        const key = el.getAttribute('data-i18n');
        if (t[key]) {
            el.innerHTML = t[key];
        }
    });

    // Translate dynamic input placeholders
    const searchBox = document.getElementById('search-box');
    if (searchBox) {
        searchBox.placeholder = t['search_placeholder'];
    }

    // Translate select filter dropdown options
    const statusSelect = document.getElementById('filter-status');
    if (statusSelect) {
        const options = statusSelect.options;
        if (options.length >= 6) {
            options[0].text = t['filter_all'];
            options[1].text = t['filter_missing_ws'];
            options[2].text = t['filter_has_ws'];
            options[3].text = t['filter_source_brain'];
            options[4].text = t['filter_source_preserved'];
            options[5].text = t['filter_source_fallback'];
        }
    }

    // Update checkbox labels in mode selector
    const modeLabels = document.querySelectorAll('.mode-selector .radio-label');
    if (modeLabels.length >= 2) {
        modeLabels[0].innerText = t['mode_auto'];
        modeLabels[1].innerText = t['mode_strict'];
    }

    // Sync select dropdown visual state
    const langSelect = document.getElementById('lang-select');
    if (langSelect) {
        langSelect.value = lang;
    }
}

// Translate Go backend event logs to display localized text if possible
function getLocalizedLog(msg) {
    const t = TRANSLATIONS[state.lang];
    if (msg.includes("Scanning conversation directories...")) {
        return t['log_scanning_backend'];
    }
    if (msg.startsWith("Found ") && msg.includes(" unique conversations")) {
        const count = msg.match(/\d+/)[0];
        return t['log_found_conv'].replace("{count}", count);
    }
    if (msg.includes("Reading metadata from database(s)...")) {
        return t['log_reading_meta'];
    }
    if (msg.startsWith("Titles: ") && msg.includes(" brain, ")) {
        const numbers = msg.match(/\d+/g);
        if (numbers && numbers.length >= 3) {
            return t['log_titles_stats']
                .replace("{brain}", numbers[0])
                .replace("{preserved}", numbers[1])
                .replace("{fallback}", numbers[2]);
        }
    }
    if (msg.includes("Starting fix process...")) {
        return t['log_start_fix'];
    }
    if (msg.startsWith("Auto-assigning workspaces for ") && msg.includes(" conversations...")) {
        const count = msg.match(/\d+/)[0];
        return t['log_auto_mapping'].replace("{count}", count);
    }
    if (msg.startsWith("Auto-assigned ") && msg.includes(" workspace(s)")) {
        const count = msg.match(/\d+/)[0];
        return t['log_auto_mapped'].replace("{count}", count);
    }
    if (msg.includes("Building final index...")) {
        return t['log_building_index'];
    }
    if (msg.startsWith("Processing ") && msg.includes("/")) {
        const numbers = msg.match(/\d+/g);
        if (numbers && numbers.length >= 2) {
            return t['log_processed']
                .replace("{step}", numbers[0])
                .replace("{total}", numbers[1]);
        }
    }
    if (msg.startsWith("Workspace: ") && msg.includes(" mapped | Timestamps injected: ")) {
        const numbers = msg.match(/\d+/g);
        if (numbers && numbers.length >= 2) {
            return t['log_workspace_summary']
                .replace("{mapped}", numbers[0])
                .replace("{injected}", numbers[1]);
        }
    }
    if (msg.includes("Writing to database(s)...")) {
        return t['log_writing_db'];
    }
    if (msg.startsWith("Updated: ")) {
        const appName = msg.replace("Updated: ", "");
        return t['log_updated_db'].replace("{appName}", appName);
    }
    if (msg.startsWith("SUCCESS! Rebuilt index with ")) {
        const count = msg.match(/\d+/)[0];
        return t['log_success'].replace("{count}", count);
    }
    if (msg.includes("Please scan conversations first")) {
        return t['err_first_scan'];
    }
    if (msg.includes("No database found")) {
        return t['err_no_db'];
    }

    return msg; // Fallback
}

// Load and Display System Paths
function loadSystemInfo() {
    const t = TRANSLATIONS[state.lang];
    GetSystemInfo()
        .then((info) => {
            state.systemInfo = info;
            
            // Set OS badge
            const osBadge = document.getElementById('os-badge');
            if (osBadge) {
                osBadge.innerText = `${t['os_badge_prefix']}${info.os}`;
            }

            // Build path diagnostics view
            const pathListEl = document.getElementById('path-list');
            if (pathListEl) {
                pathListEl.innerHTML = '';
                
                // 1. Database Check
                const dbFound = info.dbPaths && info.dbPaths.length > 0;
                pathListEl.appendChild(createPathItem(
                    dbFound ? 'sqlite-ok' : 'sqlite-missing',
                    dbFound ? 'SUCCESS' : 'WARNING',
                    dbFound 
                        ? t['lbl_sqlite_found'].replace('{count}', info.dbPaths.length)
                        : t['lbl_sqlite_missing'],
                    dbFound ? 'checkCircle' : 'alertCircle'
                ));

                // 2. Conversation Directories Check
                const convDirsFound = info.conversationDirs && info.conversationDirs.length > 0;
                pathListEl.appendChild(createPathItem(
                    convDirsFound ? 'conv-ok' : 'conv-missing',
                    convDirsFound ? 'SUCCESS' : 'WARNING',
                    convDirsFound
                        ? t['lbl_conv_found'].replace('{count}', info.conversationDirs.length)
                        : t['lbl_conv_missing'],
                    convDirsFound ? 'checkCircle' : 'alertCircle'
                ));

                // 3. Brain Directories Check
                const brainDirsFound = info.brainDirs && info.brainDirs.length > 0;
                pathListEl.appendChild(createPathItem(
                    brainDirsFound ? 'brain-ok' : 'brain-missing',
                    brainDirsFound ? 'INFO' : 'WARNING',
                    brainDirsFound
                        ? t['lbl_brain_found']
                        : t['lbl_brain_missing'],
                    brainDirsFound ? 'checkCircle' : 'alertCircle'
                ));
            }
        })
        .catch((err) => {
            console.error('Failed to get system info:', err);
        });
}

function createPathItem(id, status, text, iconKey) {
    const item = document.createElement('div');
    item.className = 'path-item';
    item.id = id;
    item.innerHTML = `
        ${ICONS[iconKey]}
        <span><strong>[${status}]</strong> ${text}</span>
    `;
    return item;
}

// Setup Event Subscriptions from Go Backend
function setupEventSubscriptions() {
    // 1. Progress event
    EventsOn('fix:progress', (data) => {
        // Update percentage and status in modal
        const percentText = document.getElementById('progress-percentage-text');
        const statusText = document.getElementById('progress-status-text');
        
        if (percentText) percentText.innerText = `${data.percent}%`;
        if (statusText) statusText.innerText = getLocalizedLog(data.message);
        
        updateProgressRing(data.percent);
    });

    // 2. Real-time log event
    EventsOn('fix:log', (data) => {
        appendTerminalLog(data.level, getLocalizedLog(data.message));
    });
}

// Update SVG Progress Circular Bar
function updateProgressRing(percent) {
    const bar = document.getElementById('progress-ring-bar');
    if (!bar) return;
    
    // Circle radius is 52, perimeter is 2 * PI * r = ~326.72
    const perimeter = 326.72;
    const offset = perimeter - (percent / 100) * perimeter;
    bar.style.strokeDashoffset = offset;
}

// Append real-time logs to the progress modal terminal
function appendTerminalLog(level, message) {
    const logBox = document.getElementById('terminal-logs');
    if (!logBox) return;

    const line = document.createElement('div');
    line.className = 'terminal-line';
    line.innerHTML = `
        <span class="term-tag ${level}">${level}</span>
        <span class="term-msg">${escapeHtml(message)}</span>
    `;
    
    logBox.appendChild(line);
    logBox.scrollTop = logBox.scrollHeight;
}

// Helper to escape HTML characters
function escapeHtml(text) {
    if (!text) return '';
    return text.toString()
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}

// Check for updates from GitHub Releases API
function runUpdateCheck() {
    CheckForUpdates()
        .then((info) => {
            if (info && info.hasUpdate) {
                const badge = document.getElementById('update-badge');
                if (badge) {
                    badge.classList.remove('hidden');
                    badge.title = `Update to v${info.latestVersion} available!`;
                    badge.innerText = `🚀 Update Available: v${info.latestVersion}`;
                }
            }
        })
        .catch((err) => {
            console.error('Update check error:', err);
        });
}

// Window Check for updates handler
window.checkUpdates = function() {
    CheckForUpdates().then((info) => {
        if (info && info.hasUpdate) {
            alert(`New version ${info.latestVersion} is available! Download from GitHub: ${info.releaseUrl}`);
        }
    });
};

// UI Step navigation updates
function updateStepNavigation(activeStep, isCompleted = false) {
    state.activeStep = activeStep;
    for (let i = 1; i <= 3; i++) {
        const navEl = document.getElementById(`step-nav-${i}`);
        if (!navEl) continue;

        navEl.classList.remove('active', 'completed');
        
        if (i === activeStep) {
            navEl.classList.add('active');
        } else if (i < activeStep) {
            navEl.classList.add('completed');
        }
    }
}

// Main Scan Conversations Routine
window.startScanning = function() {
    const t = TRANSLATIONS[state.lang];
    updateStepNavigation(2);

    // Swap views from Welcome to Scan Results
    document.getElementById('screen-welcome').classList.add('hidden');
    
    const resultsScreen = document.getElementById('screen-scan-results');
    resultsScreen.classList.remove('hidden');
    
    // Clear list to show dynamic loader
    const rowsEl = document.getElementById('conversation-rows');
    rowsEl.innerHTML = `
        <tr>
            <td colspan="5" style="text-align: center; color: var(--text-dim); padding: 40px 0;">
                <div style="margin-bottom: 12px; font-weight: 600;">${t['lbl_scanning']}</div>
            </td>
        </tr>
    `;

    // Execute scan
    ScanConversations()
        .then((result) => {
            state.conversations = result.conversations || [];
            state.filteredConversations = [...state.conversations];
            
            // Render Stats
            document.getElementById('stat-brain-count').innerText = result.stats.brain || 0;
            document.getElementById('stat-preserved-count').innerText = result.stats.preserved || 0;
            document.getElementById('stat-fallback-count').innerText = result.stats.fallback || 0;
            document.getElementById('stat-ws-count').innerText = `${result.wsCount || 0} / ${state.conversations.length}`;

            // Reset filters
            document.getElementById('search-box').value = '';
            document.getElementById('filter-status').value = 'all';
            
            // Build conversation table
            renderConversations(state.filteredConversations);
        })
        .catch((err) => {
            console.error('Scan error:', err);
            rowsEl.innerHTML = `
                <tr>
                    <td colspan="5" style="text-align: center; color: var(--accent-error); padding: 40px 0;">
                        <div style="font-weight: 700; margin-bottom: 8px;">${t['lbl_scan_failed']}</div>
                        <div style="font-size: 12px;">${escapeHtml(err)}</div>
                    </td>
                </tr>
            `;
        });
};

// Filter conversations in table
window.filterConversations = function() {
    const searchVal = document.getElementById('search-box').value.toLowerCase().trim();
    const statusVal = document.getElementById('filter-status').value;
    
    state.searchQuery = searchVal;
    state.statusFilter = statusVal;

    state.filteredConversations = state.conversations.filter((c) => {
        // Search Filter
        const matchesSearch = c.title.toLowerCase().includes(searchVal) || c.id.toLowerCase().includes(searchVal);
        
        // Status Filter
        let matchesStatus = true;
        if (statusVal === 'missing-ws') {
            matchesStatus = !c.hasWorkspace;
        } else if (statusVal === 'has-ws') {
            matchesStatus = c.hasWorkspace;
        } else if (statusVal === 'source-brain') {
            matchesStatus = c.source === 'brain';
        } else if (statusVal === 'source-preserved') {
            matchesStatus = c.source === 'preserved';
        } else if (statusVal === 'source-fallback') {
            matchesStatus = c.source === 'fallback';
        }
        
        return matchesSearch && matchesStatus;
    });

    renderConversations(state.filteredConversations);
};

// Render Conversations List in table
function renderConversations(list) {
    const t = TRANSLATIONS[state.lang];
    const rowsEl = document.getElementById('conversation-rows');
    if (!rowsEl) return;

    if (list.length === 0) {
        rowsEl.innerHTML = `
            <tr>
                <td colspan="5" style="text-align: center; color: var(--text-dim); padding: 40px 0;">
                    ${t['lbl_no_conv']}
                </td>
            </tr>
        `;
        return;
    }

    rowsEl.innerHTML = '';
    list.forEach((conv, index) => {
        const row = document.createElement('tr');
        
        // 1. Index Column
        const idxTd = document.createElement('td');
        idxTd.className = 'conv-row-index';
        idxTd.innerText = conv.index;
        
        // 2. Status Badge Column
        const statusTd = document.createElement('td');
        let statusBadge = '';
        if (conv.source === 'brain') {
            statusBadge = `<span class="status-badge-inline badge-new" title="${t['badge_new_title']}">[+]</span>`;
        } else if (conv.source === 'preserved') {
            statusBadge = `<span class="status-badge-inline badge-exist" title="${t['badge_exist_title']}">[~]</span>`;
        } else {
            statusBadge = `<span class="status-badge-inline badge-fallback" title="${t['badge_fallback_title']}">[?]</span>`;
        }
        statusTd.innerHTML = statusBadge;

        // 3. Conversation Details Column
        const detailsTd = document.createElement('td');
        detailsTd.className = 'conv-title-cell';
        detailsTd.innerHTML = `
            <div class="conv-title-text" title="${escapeHtml(conv.title)}">${escapeHtml(conv.title)}</div>
            <div class="conv-id" title="Copy full ID: ${conv.id}">${conv.id.substring(0, 16)}...</div>
        `;

        // 4. Workspace Column
        const wsTd = document.createElement('td');
        wsTd.className = 'workspace-cell';
        if (conv.hasWorkspace) {
            wsTd.innerHTML = `
                <div class="workspace-status" style="color: var(--accent-success);">
                    ${ICONS.checkCircle}
                    <span>${t['status_active']}</span>
                </div>
            `;
        } else {
            wsTd.innerHTML = `
                <div class="workspace-status" style="color: var(--accent-warning);">
                    ${ICONS.alertCircle}
                    <span>${t['status_inactive']}</span>
                </div>
            `;
        }

        // 5. Action Column
        const actionTd = document.createElement('td');
        actionTd.style.textAlign = 'right';
        
        const folderBtn = document.createElement('button');
        folderBtn.className = 'btn-action-sm';
        folderBtn.innerHTML = `
            ${ICONS.folder}
            <span>${conv.hasWorkspace ? t['btn_reassign'] : t['btn_map_dir']}</span>
        `;
        folderBtn.onclick = () => selectAndAssignWorkspace(conv.id);
        
        actionTd.appendChild(folderBtn);

        row.appendChild(idxTd);
        row.appendChild(statusTd);
        row.appendChild(detailsTd);
        row.appendChild(wsTd);
        row.appendChild(actionTd);
        
        rowsEl.appendChild(row);
    });
}

// Action: Handle native select directory and assign to workspace
function selectAndAssignWorkspace(cid) {
    const t = TRANSLATIONS[state.lang];
    SelectFolder(t['dialog_select_title'])
        .then((dirPath) => {
            if (!dirPath) return; // cancelled or empty

            AssignWorkspace(cid, dirPath)
                .then((success) => {
                    if (success) {
                        // Locate conversation item in our state and mark it
                        const cIdx = state.conversations.findIndex((c) => c.id === cid);
                        if (cIdx !== -1) {
                            state.conversations[cIdx].hasWorkspace = true;
                        }
                        
                        // Update lists and refresh
                        filterConversations();
                        
                        // Recalculate workspace stats
                        const totalMapped = state.conversations.filter(c => c.hasWorkspace).length;
                        document.getElementById('stat-ws-count').innerText = `${totalMapped} / ${state.conversations.length}`;
                        
                        const logMsg = t['log_manual_assign'].replace("{cid}", cid.substring(0, 8)).replace("{dirPath}", dirPath);
                        appendTerminalLog('info', logMsg);
                    }
                })
                .catch((err) => {
                    alert(`Failed to assign workspace path: ${err}`);
                });
        })
        .catch((err) => {
            console.error('Select folder dialog error:', err);
        });
}

// Executing Trajectory DB Index fixes
window.executeFix = function() {
    const t = TRANSLATIONS[state.lang];
    updateStepNavigation(3);
    
    // Select selected mode
    let selectedMode = 'auto';
    const modes = document.getElementsByName('fix-mode');
    for (let m of modes) {
        if (m.checked) {
            selectedMode = m.value;
            break;
        }
    }

    // Toggle modal visibility
    const modalEl = document.getElementById('overlay-progress');
    modalEl.classList.remove('hidden');

    // Reset progress UI and status
    document.getElementById('progress-title').innerText = t['progress_title'];
    document.getElementById('progress-status-text').innerText = t['progress_status'];
    document.getElementById('progress-percentage-text').innerText = '0%';
    updateProgressRing(0);

    // Clear Terminal Logs
    const logBox = document.getElementById('terminal-logs');
    logBox.innerHTML = '';
    appendTerminalLog('sys', t['log_initialized']);

    // Hide summary panel and show loader during process
    document.getElementById('success-summary').classList.add('hidden');
    document.getElementById('btn-dismiss-modal').classList.add('hidden');
    document.getElementById('terminal-logs').classList.remove('hidden');

    // Execute Wails bound fix function
    RunFix({ mode: selectedMode })
        .then((result) => {
            if (result.success) {
                // Done! Update UI to success state
                document.getElementById('progress-title').innerText = t['success_title'];
                document.getElementById('progress-status-text').innerText = t['success_desc'];
                document.getElementById('progress-percentage-text').innerText = '100%';
                updateProgressRing(100);

                // Show statistics panel
                document.getElementById('summary-total').innerText = result.total;
                document.getElementById('summary-ws-mapped').innerText = result.workspaceMapped;
                document.getElementById('summary-ts-injected').innerText = result.timestampsInjected;

                // Build DB write details
                const dbResultsEl = document.getElementById('summary-db-results');
                dbResultsEl.innerHTML = '';
                if (result.dbResults && result.dbResults.length > 0) {
                    result.dbResults.forEach((db) => {
                        const item = document.createElement('div');
                        item.className = 'db-item-row';
                        item.innerHTML = `
                            <span><strong>${escapeHtml(db.appName)}</strong>: Rebuilt Index</span>
                            <span class="db-bk" title="Backup location: ${escapeHtml(db.backupFile)}">Backup: ${escapeHtml(db.backupFile.split('/').pop())}</span>
                        `;
                        dbResultsEl.appendChild(item);
                    });
                } else {
                    dbResultsEl.innerHTML = `<div style="text-align: center; color: var(--text-dim); font-size: 12px;">${t['log_db_no_change']}</div>`;
                }

                // Show/Hide Panels
                document.getElementById('terminal-logs').classList.add('hidden');
                document.getElementById('success-summary').classList.remove('hidden');
                document.getElementById('btn-dismiss-modal').classList.remove('hidden');
            } else {
                handleFixError(result.error || 'Unknown backend compilation or disk write failure.');
            }
        })
        .catch((err) => {
            handleFixError(err);
        });
};

function handleFixError(errorText) {
    const t = TRANSLATIONS[state.lang];
    document.getElementById('progress-title').innerText = t['fix_failed'];
    document.getElementById('progress-status-text').innerText = t['fix_failed_desc'];
    appendTerminalLog('error', `Fix process failed: ${errorText}`);
    
    const btnDismiss = document.getElementById('btn-dismiss-modal');
    const dismissText = document.getElementById('btn-dismiss-text');
    if (dismissText) dismissText.innerText = t['btn_back_dashboard'];
    btnDismiss.classList.remove('hidden');
}

// Close Progress Modal back to dashboard view
window.closeProgressModal = function() {
    // Hide modal overlay
    const modalEl = document.getElementById('overlay-progress');
    modalEl.classList.add('hidden');

    // Trigger step navigation back to scan stage
    updateStepNavigation(2);
};
