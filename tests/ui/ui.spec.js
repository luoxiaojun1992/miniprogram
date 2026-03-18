// @ts-check
const { test, expect } = require('@playwright/test');

/**
 * Playwright UI automation tests for miniprogram.
 *
 * Environment variables:
 *   APP_BASE_URL  – backend API   (default http://localhost:8080)
 *   UI_BASE_URL   – UI server     (default http://localhost:8081)
 */

const APP_BASE_URL = process.env.APP_BASE_URL || 'http://localhost:8080';
const UI_BASE_URL  = process.env.UI_BASE_URL  || 'http://localhost:8081';

/* ─── helpers ─────────────────────────────────────────────────────── */

/** Obtain an admin JWT token via the debug endpoint. */
async function getAdminToken(request) {
  const res = await request.post(`${APP_BASE_URL}/v1/debug/token`, {
    data: { user_id: 1, user_type: 3 },
  });
  const body = await res.json();
  return body.data?.access_token ?? '';
}

/** Obtain a regular-user JWT token via the debug endpoint. */
async function getUserToken(request) {
  const res = await request.post(`${APP_BASE_URL}/v1/debug/token`, {
    data: { user_id: 2, user_type: 1 },
  });
  const body = await res.json();
  return body.data?.access_token ?? '';
}

async function attachJSON(testInfo, name, payload) {
  await testInfo.attach(name, {
    body: Buffer.from(JSON.stringify(payload, null, 2), 'utf-8'),
    contentType: 'application/json',
  });
}

async function ensureModuleIDForArticle(request, adminToken) {
  const headers = { Authorization: `Bearer ${adminToken}` };
  const listRes = await request.get(`${APP_BASE_URL}/v1/admin/modules?page=1&page_size=20`, { headers });
  if (listRes.ok()) {
    const listBody = await listRes.json();
    const modules = listBody.data?.list ?? listBody.data ?? [];
    if (Array.isArray(modules)) {
      const existing = modules.find((item) => Number(item?.id) > 0);
      if (existing) {
        return Number(existing.id);
      }
    }
  }

  const createRes = await request.post(`${APP_BASE_URL}/v1/admin/modules`, {
    headers,
    data: {
      title: `UI Report Module ${Date.now()}-${Math.floor(Math.random() * 100000)}`,
      description: 'ui-report-module',
      sort_order: 0,
    },
  });
  expect(createRes.ok()).toBeTruthy();
  const createBody = await createRes.json();
  const moduleID = createBody.data?.id ?? 0;
  expect(moduleID).toBeGreaterThan(0);
  return moduleID;
}

/* ═══════════════════════════════════════════════════════════════════ */
/*  ADMIN PORTAL TESTS                                                */
/* ═══════════════════════════════════════════════════════════════════ */

test.describe('Admin Portal', () => {

  test.describe('Login Page', () => {

    test('shows login form on first visit', async ({ page }) => {
      await page.goto(`${UI_BASE_URL}/admin/index.html`);
      await expect(page.locator('h1')).toContainText('知识库管理后台');
      await expect(page.locator('input[type="email"], input[placeholder*="admin"]')).toBeVisible();
      await expect(page.getByRole('button', { name: /登.*录/ })).toBeVisible();
    });

    test('validates empty fields', async ({ page }) => {
      await page.goto(`${UI_BASE_URL}/admin/index.html`);
      await page.getByRole('button', { name: /登.*录/ }).click();
      // Should show validation error – either an error-text element or the inputs remain visible
      await expect(page.locator('.login-card')).toBeVisible();
    });

    test('login with valid credentials and navigate dashboard', async ({ page }) => {
      await page.goto(`${UI_BASE_URL}/admin/index.html`);
      // Set API base URL
      await page.locator('input[placeholder*="localhost"]').first().fill(`${APP_BASE_URL}/v1`);
      await page.locator('input[type="email"], input[placeholder*="admin"]').fill('admin@example.com');
      await page.locator('input[type="password"]').fill('Test@123456');
      await page.getByRole('button', { name: /登.*录/ }).click();

      // Should land on dashboard
      await expect(page.locator('.sidebar').first()).toBeVisible({ timeout: 15000 });
      await expect(page.locator('h3, .page-title, header').first()).toContainText(/仪表盘|Dashboard/i);
    });
  });

  test.describe('Authenticated Pages', () => {
    // Pre-inject token to bypass login for each test
    test.beforeEach(async ({ page, request }) => {
      const token = await getAdminToken(request);
      await page.goto(`${UI_BASE_URL}/admin/index.html`);
      await page.evaluate(([t, base]) => {
        localStorage.setItem('admin_token', t);
        localStorage.setItem('admin_api_base', base + '/v1');
      }, [token, APP_BASE_URL]);
      await page.reload();
      // Wait for sidebar to appear (indicates successful auth)
      await expect(page.locator('.sidebar, nav').first()).toBeVisible({ timeout: 15000 });
    });

    test('dashboard shows statistics cards', async ({ page }) => {
      await expect(page.locator('.stats-grid, .stat-card').first()).toBeVisible();
      // Verify 4 stat cards
      const cards = page.locator('.stat-card');
      await expect(cards).toHaveCount(4);
      await expect(page.getByText('用户总数')).toBeVisible();
      await expect(page.getByText('文章总数')).toBeVisible();
      await expect(page.getByText('课程总数')).toBeVisible();
      await expect(page.getByText('评论总数')).toBeVisible();
    });

    test('navigate to user management', async ({ page }) => {
      await page.getByText('用户管理').click();
      await expect(page.locator('h3, .page-title').first()).toContainText(/用户管理/);
      // Table headers visible
      await expect(page.getByRole('cell', { name: 'ID' }).or(page.getByText('ID').first())).toBeVisible();
      // "新增管理员" button visible
      await expect(page.getByRole('button', { name: /新增/ })).toBeVisible();
    });

    test('navigate to attribute management', async ({ page }) => {
      await page.getByText('属性管理').click();
      await expect(page.locator('h3, .page-title').first()).toContainText(/属性管理/);
      await expect(page.getByRole('button', { name: /新增属性/ })).toBeVisible();
    });

    test('navigate to article management', async ({ page }) => {
      await page.getByText('文章管理').click();
      await expect(page.locator('h3, .page-title').first()).toContainText(/文章管理/);
      await expect(page.getByRole('button', { name: /新增文章/ })).toBeVisible();
    });

    test('open create article modal', async ({ page }) => {
      await page.getByText('文章管理').click();
      await page.getByRole('button', { name: /新增文章/ }).click();
      // Modal should appear with form fields
      await expect(page.locator('.modal-overlay, .modal')).toBeVisible();
      await expect(page.getByText('标题').first()).toBeVisible();
      await expect(page.getByText('内容').first()).toBeVisible();
    });

    test('report includes created content, created content list, and uploaded file', async ({ page, request }, testInfo) => {
      const adminToken = await getAdminToken(request);
      const uniqueTitle = `UI Report Article ${Date.now()}-${Math.floor(Math.random() * 100000)}`;
      const moduleID = await ensureModuleIDForArticle(request, adminToken);

      let articleId = 0;
      await test.step('创建内容：创建文章', async () => {
        const articleRes = await request.post(`${APP_BASE_URL}/v1/admin/articles`, {
          headers: { Authorization: `Bearer ${adminToken}` },
          data: { title: uniqueTitle, summary: 'ui-report-summary', content: 'ui-report-content', content_type: 1, module_id: moduleID },
        });
        expect(articleRes.ok()).toBeTruthy();
        const articleBody = await articleRes.json();
        articleId = articleBody.data?.id ?? 0;
        expect(articleId).toBeGreaterThan(0);
        await attachJSON(testInfo, 'created-content.json', articleBody);
      });

      await test.step('创建后的内容列表：接口与页面可见', async () => {
        const listRes = await request.get(
          `${APP_BASE_URL}/v1/admin/articles?page=1&page_size=20&keyword=${encodeURIComponent(uniqueTitle)}`,
          { headers: { Authorization: `Bearer ${adminToken}` } },
        );
        expect(listRes.ok()).toBeTruthy();
        const listBody = await listRes.json();
        const list = listBody.data?.list ?? [];
        expect(Array.isArray(list)).toBeTruthy();
        expect(list.some((item) => item.id === articleId || item.title === uniqueTitle)).toBeTruthy();
        await attachJSON(testInfo, 'created-content-list.json', listBody);

        await page.getByText('文章管理').click();
        await expect(page.locator('h3, .page-title').first()).toContainText(/文章管理/);
        await page.locator('.search-input').first().fill(uniqueTitle);
        await page.keyboard.press('Enter');
        await expect(page.getByText(uniqueTitle).first()).toBeVisible({ timeout: 15000 });
      });

      await test.step('上传成功的文件：预签名上传并验证文件ID', async () => {
        const filename = `ui-report-${Date.now()}.png`;
        const presignRes = await request.get(
          `${APP_BASE_URL}/v1/admin/upload/files/presign?filename=${encodeURIComponent(filename)}&usage=embedded&expires_in=600`,
          { headers: { Authorization: `Bearer ${adminToken}` } },
        );
        expect(presignRes.ok()).toBeTruthy();
        const presignBody = await presignRes.json();
        const fileID = presignBody.data?.file_id ?? 0;
        const putURL = presignBody.data?.put_url ?? '';
        expect(fileID).toBeGreaterThan(0);
        expect(typeof putURL).toBe('string');
        expect(putURL.length).toBeGreaterThan(0);

        const uploadRes = await request.fetch(putURL, {
          method: 'PUT',
          headers: { 'Content-Type': 'image/png' },
          data: Buffer.from('ui-report-upload-png'),
        });
        expect(uploadRes.ok()).toBeTruthy();
        const uploadBody = await uploadRes.text();
        await attachJSON(testInfo, 'uploaded-file.json', {
          file_id: fileID,
          key: presignBody.data?.key,
          static_url: presignBody.data?.static_url ?? '',
          upload_status: uploadRes.status(),
          upload_response: uploadBody,
        });
      });
    });

    test('navigate to course management', async ({ page }) => {
      await page.getByText('课程管理').click();
      await expect(page.locator('h3, .page-title').first()).toContainText(/课程管理/);
      await expect(page.getByRole('button', { name: /新增课程/ })).toBeVisible();
    });

    test('navigate to module management', async ({ page }) => {
      await page.getByText('模块管理').click();
      await expect(page.locator('h3, .page-title').first()).toContainText(/模块管理/);
    });

    test('navigate to banner management', async ({ page }) => {
      await page.getByText('轮播图管理').click();
      await expect(page.locator('h3, .page-title').first()).toContainText(/轮播图管理/);
      await expect(page.getByRole('button', { name: /新增轮播图/ })).toBeVisible();
    });

    test('navigate to comment management', async ({ page }) => {
      // Expand 互动管理 section (collapsed by default)
      await page.locator('.sidebar-menu .menu-item').filter({ hasText: '互动管理' }).click();
      await page.locator('.submenu .menu-item').filter({ hasText: '评论管理' }).click();
      await expect(page.locator('h3, .page-title').first()).toContainText(/评论管理|评论/);
    });

    test('navigate to role management', async ({ page }) => {
      // Expand 权限管理 section (collapsed by default)
      await page.locator('.sidebar-menu .menu-item').filter({ hasText: '权限管理' }).click();
      await page.locator('.submenu .menu-item').filter({ hasText: '角色管理' }).click();
      await expect(page.locator('h3, .page-title').first()).toContainText(/角色管理/);
    });

    test('navigate to permission tree', async ({ page }) => {
      // Expand 权限管理 section (collapsed by default)
      await page.locator('.sidebar-menu .menu-item').filter({ hasText: '权限管理' }).click();
      await page.locator('.submenu .menu-item').filter({ hasText: '权限列表' }).click();
      await expect(page.locator('h3, .page-title').first()).toContainText(/权限/);
    });

    test('navigate to wechat config', async ({ page }) => {
      // Expand 系统管理 section (collapsed by default)
      await page.locator('.sidebar-menu .menu-item').filter({ hasText: '系统管理' }).click();
      await page.locator('.submenu .menu-item').filter({ hasText: '微信配置' }).click();
      await expect(page.locator('h3, .page-title').first()).toContainText(/微信配置/);
    });

    test('navigate to log config', async ({ page }) => {
      // Expand 系统管理 section (collapsed by default)
      await page.locator('.sidebar-menu .menu-item').filter({ hasText: '系统管理' }).click();
      await page.locator('.submenu .menu-item').filter({ hasText: '日志配置' }).click();
      await expect(page.locator('h3, .page-title').first()).toContainText(/日志配置/);
    });

    test('navigate to audit logs', async ({ page }) => {
      // Expand 系统管理 section (collapsed by default)
      await page.locator('.sidebar-menu .menu-item').filter({ hasText: '系统管理' }).click();
      await page.locator('.submenu .menu-item').filter({ hasText: '审计日志' }).click();
      await expect(page.locator('h3, .page-title').first()).toContainText(/审计日志/);
    });

    test('admin write action appears in audit logs', async ({ page }) => {
      await page.locator('.sidebar-menu .menu-item').filter({ hasText: '内容管理' }).click();
      await page.locator('.submenu .menu-item').filter({ hasText: '模块管理' }).click();
      await expect(page.locator('h3, .page-title').first()).toContainText(/模块管理/);

      await page.getByRole('button', { name: /新增模块/ }).click();
      await page.locator('.modal-body input.form-control').first().fill(`UI Audit Module ${Date.now()}-${Math.floor(Math.random() * 100000)}`);
      await page.getByRole('button', { name: /^保存$/ }).click();

      await page.locator('.sidebar-menu .menu-item').filter({ hasText: '系统管理' }).click();
      await page.locator('.submenu .menu-item').filter({ hasText: '审计日志' }).click();
      await expect(page.locator('h3, .page-title').first()).toContainText(/审计日志/);
      await expect(page.getByText('modules').first()).toBeVisible({ timeout: 15000 });
      await expect(page.getByText('create').first()).toBeVisible();
    });

    test('logout returns to login page', async ({ page }) => {
      await page.getByRole('button', { name: /退出登录/ }).click();
      await expect(page.locator('h1')).toContainText('知识库管理后台');
    });
  });
});

/* ═══════════════════════════════════════════════════════════════════ */
/*  MINIPROGRAM SIMULATOR TESTS                                       */
/* ═══════════════════════════════════════════════════════════════════ */

test.describe('Miniprogram Simulator', () => {

  test.describe('Login Page', () => {

    test('shows login interface on first visit', async ({ page }) => {
      await page.goto(`${UI_BASE_URL}/miniprogram/index.html`);
      await expect(page.getByText('小程序模拟器').first()).toBeVisible();
      await expect(page.getByRole('button', { name: /微信模拟登录/ })).toBeVisible();
      await expect(page.getByText('调试 Token')).toBeVisible();
    });

    test('debug token button disabled when empty', async ({ page }) => {
      await page.goto(`${UI_BASE_URL}/miniprogram/index.html`);
      await expect(page.getByRole('button', { name: /使用调试Token/ })).toBeDisabled();
    });

    test('debug token button enabled after input', async ({ page }) => {
      await page.goto(`${UI_BASE_URL}/miniprogram/index.html`);
      await page.locator('input[placeholder*="调试"]').fill('some-token');
      await expect(page.getByRole('button', { name: /使用调试Token/ })).toBeEnabled();
    });
  });

  test.describe('Authenticated Pages', () => {
    // Pre-inject a user token via localStorage
    test.beforeEach(async ({ page, request }) => {
      const token = await getUserToken(request);
      await page.goto(`${UI_BASE_URL}/miniprogram/index.html`);
      await page.evaluate(([t, base]) => {
        localStorage.setItem('mp_token', t);
        localStorage.setItem('mp_api_base', base + '/v1');
        localStorage.setItem('mp_user', JSON.stringify({ nickname: 'Test User' }));
      }, [token, APP_BASE_URL]);
      await page.reload();
      // Wait for tab bar to appear (indicates authenticated state)
      await expect(page.getByText('首页').first()).toBeVisible({ timeout: 15000 });
    });

    test('home page shows sections', async ({ page }) => {
      await expect(page.getByText('功能模块').first()).toBeVisible();
      await expect(page.getByText('推荐文章').first()).toBeVisible();
    });

    test('tab navigation to articles', async ({ page }) => {
      // Click the articles tab
      await page.locator('.tab-item').filter({ hasText: '文章' }).click();
      // v-show keeps all tab DOMs alive; target the visible articles page title
      await expect(page.locator('.page-title').filter({ hasText: '文章' })).toBeVisible();
      await expect(page.locator('input[placeholder*="搜索文章"]')).toBeVisible();
    });

    test('tab navigation to courses', async ({ page }) => {
      await page.locator('.tab-item').filter({ hasText: '课程' }).click();
      await expect(page.getByText('课程').first()).toBeVisible();
    });

    test('tab navigation to profile', async ({ page }) => {
      await page.locator('.tab-item').filter({ hasText: '我的' }).click();
      await expect(page.getByText('我的收藏')).toBeVisible();
      await expect(page.getByText('学习记录')).toBeVisible();
      await expect(page.getByText('我的通知')).toBeVisible();
      await expect(page.getByRole('button', { name: /退出登录/ })).toBeVisible();
    });

    test('logout returns to login page', async ({ page }) => {
      await page.locator('.tab-item').filter({ hasText: '我的' }).click();
      await page.getByRole('button', { name: /退出登录/ }).click();
      await expect(page.getByText('小程序模拟器').first()).toBeVisible();
      await expect(page.getByRole('button', { name: /微信模拟登录/ })).toBeVisible();
    });

    test('notifications page shows like/comment notifications after interactions', async ({ page, request }) => {
      const adminTokenRes = await request.post(`${APP_BASE_URL}/v1/debug/token`, { data: { user_id: 1, user_type: 3 } });
      const adminBody = await adminTokenRes.json();
      const adminToken = adminBody.data?.access_token ?? '';
      const userToken = await getUserToken(request);

      const articleRes = await request.post(`${APP_BASE_URL}/v1/admin/articles`, {
        headers: { Authorization: `Bearer ${adminToken}` },
        data: { title: `UI Notif ${Date.now()}`, summary: 's', content: 'c', content_type: 1, module_id: 1 },
      });
      const articleBody = await articleRes.json();
      const articleId = articleBody.data?.id;
      await request.post(`${APP_BASE_URL}/v1/admin/articles/${articleId}/publish`, {
        headers: { Authorization: `Bearer ${adminToken}` },
        data: { status: 1 },
      });
      await request.post(`${APP_BASE_URL}/v1/likes/1/${articleId}`, { headers: { Authorization: `Bearer ${userToken}` } });
      await request.post(`${APP_BASE_URL}/v1/comments/1/${articleId}`, {
        headers: { Authorization: `Bearer ${userToken}` },
        data: { content: `ui-comment-${Date.now()}` },
      });

      await page.goto(`${UI_BASE_URL}/miniprogram/index.html`);
      await page.evaluate(([t, base]) => {
        localStorage.setItem('mp_token', t);
        localStorage.setItem('mp_api_base', base + '/v1');
        localStorage.setItem('mp_user', JSON.stringify({ nickname: 'Admin User' }));
      }, [adminToken, APP_BASE_URL]);
      await page.reload();
      await page.locator('.tab-item').filter({ hasText: '我的' }).click();
      await page.getByText('我的通知').click();
      await expect(page.getByText(/收到新的点赞|收到新的评论/).first()).toBeVisible({ timeout: 15000 });
    });
  });
});
