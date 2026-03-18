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

function escapeRegExp(value) {
  return String(value).replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
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

    test('admin CRUD operations for all resource types', async ({ page, request }, testInfo) => {
      const adminToken = await getAdminToken(request);
      const userToken = await getUserToken(request);
      const headers = { Authorization: `Bearer ${adminToken}` };
      const unique = `${Date.now()}-${testInfo.retry}-${testInfo.workerIndex}`;

      const moduleTitle = `UI CRUD Module ${unique}`;
      const attributeName = `UI CRUD Attribute ${unique}`;
      const userEmail = `ui-crud-${unique}@example.com`;
      const userNickname = `UI CRUD User ${unique}`;
      const roleName = `UI CRUD Role ${unique}`;
      const bannerTitle = `UI CRUD Banner ${unique}`;
      const articleTitle = `UI CRUD Article ${unique}`;
      const commentText = `ui-crud-comment-${unique}`;
      const courseTitle = `UI CRUD Course ${unique}`;
      const unitTitle = `UI CRUD Unit ${unique}`;
      const pageTitle = `UI CRUD Page ${unique}`;
      const MODULE_PAGE_CONTENT_TYPE = 1;
      const UNIT_DURATION_INITIAL = 10;
      const UNIT_DURATION_UPDATED = 20;

      // 1) module CRUD
      const moduleCreateRes = await request.post(`${APP_BASE_URL}/v1/admin/modules`, {
        headers,
        data: { title: moduleTitle, description: 'ui-crud-module', sort_order: 0 },
      });
      expect(moduleCreateRes.ok()).toBeTruthy();
      const moduleCreateBody = await moduleCreateRes.json();
      const moduleID = Number(moduleCreateBody.data?.id || 0);
      expect(moduleID).toBeGreaterThan(0);
      const moduleListAfterCreate = await request.get(
        `${APP_BASE_URL}/v1/modules?page=1&page_size=100&keyword=${encodeURIComponent(moduleTitle)}`,
        { headers },
      );
      const moduleListAfterCreateBody = await moduleListAfterCreate.json();
      expect((moduleListAfterCreateBody.data?.list || moduleListAfterCreateBody.data || []).some((item) => item.id === moduleID)).toBeTruthy();
      await page.getByText('模块管理').click();
      await expect(page.locator('h3, .page-title').first()).toContainText(/模块管理/);

      const moduleUpdatedTitle = `${moduleTitle} Updated`;
      const moduleUpdateRes = await request.put(`${APP_BASE_URL}/v1/admin/modules/${moduleID}`, {
        headers,
        data: { title: moduleUpdatedTitle, description: 'ui-crud-module-updated', sort_order: 1 },
      });
      expect(moduleUpdateRes.ok()).toBeTruthy();
      const moduleListAfterUpdate = await request.get(
        `${APP_BASE_URL}/v1/modules?page=1&page_size=100&keyword=${encodeURIComponent(moduleUpdatedTitle)}`,
        { headers },
      );
      const moduleListAfterUpdateBody = await moduleListAfterUpdate.json();
      expect((moduleListAfterUpdateBody.data?.list || moduleListAfterUpdateBody.data || []).some((item) => item.id === moduleID)).toBeTruthy();

      // 2) module pages CRUD
      const pageCreateRes = await request.post(`${APP_BASE_URL}/v1/admin/modules/${moduleID}/pages`, {
        headers,
        data: { title: pageTitle, content: 'ui-crud-page-content', content_type: MODULE_PAGE_CONTENT_TYPE, sort_order: 0 },
      });
      expect(pageCreateRes.ok()).toBeTruthy();
      const pageCreateBody = await pageCreateRes.json();
      const pageID = Number(pageCreateBody.data?.id || 0);
      expect(pageID).toBeGreaterThan(0);
      const pageListAfterCreate = await request.get(`${APP_BASE_URL}/v1/admin/modules/${moduleID}/pages`, { headers });
      const pageListAfterCreateBody = await pageListAfterCreate.json();
      expect((pageListAfterCreateBody.data?.list || pageListAfterCreateBody.data || []).some((item) => item.id === pageID)).toBeTruthy();
      const pageUpdatedTitle = `${pageTitle} Updated`;
      const pageUpdateRes = await request.put(`${APP_BASE_URL}/v1/admin/modules/${moduleID}/pages/${pageID}`, {
        headers,
        data: { title: pageUpdatedTitle, content: 'ui-crud-page-content-updated', content_type: MODULE_PAGE_CONTENT_TYPE, sort_order: 2 },
      });
      expect(pageUpdateRes.ok()).toBeTruthy();
      const pageListAfterUpdate = await request.get(`${APP_BASE_URL}/v1/admin/modules/${moduleID}/pages`, { headers });
      const pageListAfterUpdateBody = await pageListAfterUpdate.json();
      expect((pageListAfterUpdateBody.data?.list || pageListAfterUpdateBody.data || []).some((item) => item.id === pageID && item.title === pageUpdatedTitle)).toBeTruthy();
      const pageDeleteRes = await request.delete(`${APP_BASE_URL}/v1/admin/modules/${moduleID}/pages/${pageID}`, { headers });
      expect(pageDeleteRes.ok()).toBeTruthy();
      const pageListAfterDelete = await request.get(`${APP_BASE_URL}/v1/admin/modules/${moduleID}/pages`, { headers });
      const pageListAfterDeleteBody = await pageListAfterDelete.json();
      expect((pageListAfterDeleteBody.data?.list || pageListAfterDeleteBody.data || []).some((item) => item.id === pageID)).toBeFalsy();

      // 3) attribute CRUD
      const attrCreateRes = await request.post(`${APP_BASE_URL}/v1/admin/attributes`, {
        headers,
        data: { name: attributeName },
      });
      expect(attrCreateRes.ok()).toBeTruthy();
      const attrCreateBody = await attrCreateRes.json();
      const attributeID = Number(attrCreateBody.data?.id || 0);
      expect(attributeID).toBeGreaterThan(0);
      const attrListAfterCreate = await request.get(`${APP_BASE_URL}/v1/admin/attributes`, { headers });
      const attrListAfterCreateBody = await attrListAfterCreate.json();
      expect((attrListAfterCreateBody.data || []).some((item) => item.id === attributeID)).toBeTruthy();
      await page.getByText('属性管理').click();
      await expect(page.getByText(attributeName).first()).toBeVisible({ timeout: 15000 });
      const attributeUpdatedName = `${attributeName} Updated`;
      const attrUpdateRes = await request.put(`${APP_BASE_URL}/v1/admin/attributes/${attributeID}`, {
        headers,
        data: { name: attributeUpdatedName },
      });
      expect(attrUpdateRes.ok()).toBeTruthy();
      const attrListAfterUpdate = await request.get(`${APP_BASE_URL}/v1/admin/attributes`, { headers });
      const attrListAfterUpdateBody = await attrListAfterUpdate.json();
      expect((attrListAfterUpdateBody.data || []).some((item) => item.id === attributeID && item.name === attributeUpdatedName)).toBeTruthy();
      const attrDeleteRes = await request.delete(`${APP_BASE_URL}/v1/admin/attributes/${attributeID}`, { headers });
      expect(attrDeleteRes.ok()).toBeTruthy();
      const attrListAfterDelete = await request.get(`${APP_BASE_URL}/v1/admin/attributes`, { headers });
      const attrListAfterDeleteBody = await attrListAfterDelete.json();
      expect((attrListAfterDeleteBody.data || []).some((item) => item.id === attributeID)).toBeFalsy();

      // 4) user CRUD
      const userCreateRes = await request.post(`${APP_BASE_URL}/v1/admin/users`, {
        headers,
        data: { email: userEmail, password: 'Test@123456', nickname: userNickname, user_type: 2 },
      });
      expect(userCreateRes.ok()).toBeTruthy();
      const userCreateBody = await userCreateRes.json();
      const userID = Number(userCreateBody.data?.id || 0);
      expect(userID).toBeGreaterThan(0);
      const userListAfterCreate = await request.get(
        `${APP_BASE_URL}/v1/admin/users?page=1&page_size=100&keyword=${encodeURIComponent(userNickname)}`,
        { headers },
      );
      const userListAfterCreateBody = await userListAfterCreate.json();
      expect((userListAfterCreateBody.data?.list || []).some((item) => item.id === userID)).toBeTruthy();
      await page.getByText('用户管理').click();
      await page.locator('.search-input').first().fill(userNickname);
      await page.keyboard.press('Enter');
      await expect(page.getByText(userNickname).first()).toBeVisible({ timeout: 15000 });
      const updatedNickname = `UI CRUD User Updated ${unique}`;
      const userUpdateRes = await request.put(`${APP_BASE_URL}/v1/admin/users/${userID}`, {
        headers,
        data: { nickname: updatedNickname, user_type: 2, status: 1 },
      });
      expect(userUpdateRes.ok()).toBeTruthy();
      const userListAfterUpdate = await request.get(
        `${APP_BASE_URL}/v1/admin/users?page=1&page_size=100&keyword=${encodeURIComponent(updatedNickname)}`,
        { headers },
      );
      const userListAfterUpdateBody = await userListAfterUpdate.json();
      expect((userListAfterUpdateBody.data?.list || []).some((item) => item.id === userID)).toBeTruthy();
      const userDeleteRes = await request.delete(`${APP_BASE_URL}/v1/admin/users/${userID}`, { headers });
      expect(userDeleteRes.ok()).toBeFalsy();

      // 5) role CRUD
      const roleCreateRes = await request.post(`${APP_BASE_URL}/v1/admin/roles`, {
        headers,
        data: { name: roleName, description: 'ui-crud-role', parent_id: 0, permission_ids: [] },
      });
      expect(roleCreateRes.ok()).toBeTruthy();
      const roleCreateBody = await roleCreateRes.json();
      const roleID = Number(roleCreateBody.data?.id || 0);
      expect(roleID).toBeGreaterThan(0);
      const roleListAfterCreate = await request.get(`${APP_BASE_URL}/v1/admin/roles?page=1&page_size=100`, { headers });
      const roleListAfterCreateBody = await roleListAfterCreate.json();
      expect((roleListAfterCreateBody.data?.list || roleListAfterCreateBody.data || []).some((item) => item.id === roleID)).toBeTruthy();
      await page.locator('.sidebar-menu .menu-item').filter({ hasText: '权限管理' }).click();
      await page.locator('.submenu .menu-item').filter({ hasText: '角色管理' }).click();
      await expect(page.getByText(roleName).first()).toBeVisible({ timeout: 15000 });
      const roleUpdatedName = `${roleName} Updated`;
      const roleUpdateRes = await request.put(`${APP_BASE_URL}/v1/admin/roles/${roleID}`, {
        headers,
        data: { name: roleUpdatedName, description: 'ui-crud-role-updated', parent_id: 0, permission_ids: [] },
      });
      expect(roleUpdateRes.ok()).toBeTruthy();
      const roleListAfterUpdate = await request.get(`${APP_BASE_URL}/v1/admin/roles?page=1&page_size=100`, { headers });
      const roleListAfterUpdateBody = await roleListAfterUpdate.json();
      expect((roleListAfterUpdateBody.data?.list || roleListAfterUpdateBody.data || []).some((item) => item.id === roleID && item.name === roleUpdatedName)).toBeTruthy();
      const roleDeleteRes = await request.delete(`${APP_BASE_URL}/v1/admin/roles/${roleID}`, { headers });
      expect(roleDeleteRes.ok()).toBeTruthy();
      const roleListAfterDelete = await request.get(`${APP_BASE_URL}/v1/admin/roles?page=1&page_size=100`, { headers });
      const roleListAfterDeleteBody = await roleListAfterDelete.json();
      expect((roleListAfterDeleteBody.data?.list || roleListAfterDeleteBody.data || []).some((item) => item.id === roleID)).toBeFalsy();

      // 6) article + comment CRUD
      const articleCreateRes = await request.post(`${APP_BASE_URL}/v1/admin/articles`, {
        headers,
        data: { title: articleTitle, summary: 'ui-crud-article', content: 'ui-crud-content', content_type: 1, module_id: moduleID },
      });
      expect(articleCreateRes.ok()).toBeTruthy();
      const articleCreateBody = await articleCreateRes.json();
      const articleID = Number(articleCreateBody.data?.id || 0);
      expect(articleID).toBeGreaterThan(0);
      const articleListAfterCreate = await request.get(
        `${APP_BASE_URL}/v1/admin/articles?page=1&page_size=100&keyword=${encodeURIComponent(articleTitle)}`,
        { headers },
      );
      const articleListAfterCreateBody = await articleListAfterCreate.json();
      expect((articleListAfterCreateBody.data?.list || []).some((item) => item.id === articleID)).toBeTruthy();
      await page.getByText('文章管理').click();
      await page.locator('.search-input').first().fill(articleTitle);
      await page.keyboard.press('Enter');
      await expect(page.getByText(articleTitle).first()).toBeVisible({ timeout: 15000 });
      const articleUpdatedTitle = `${articleTitle} Updated`;
      const articleUpdateRes = await request.put(`${APP_BASE_URL}/v1/admin/articles/${articleID}`, {
        headers,
        data: { title: articleUpdatedTitle, summary: 'ui-crud-article-updated', content: 'ui-crud-content-updated', content_type: 1, module_id: moduleID },
      });
      expect(articleUpdateRes.ok()).toBeTruthy();
      const articleListAfterUpdate = await request.get(
        `${APP_BASE_URL}/v1/admin/articles?page=1&page_size=100&keyword=${encodeURIComponent(articleUpdatedTitle)}`,
        { headers },
      );
      const articleListAfterUpdateBody = await articleListAfterUpdate.json();
      expect((articleListAfterUpdateBody.data?.list || []).some((item) => item.id === articleID)).toBeTruthy();
      const publishRes = await request.post(`${APP_BASE_URL}/v1/admin/articles/${articleID}/publish`, {
        headers,
        data: { status: 1 },
      });
      expect(publishRes.ok()).toBeTruthy();
      const createCommentRes = await request.post(`${APP_BASE_URL}/v1/comments/1/${articleID}`, {
        headers: { Authorization: `Bearer ${userToken}` },
        data: { content: commentText },
      });
      expect(createCommentRes.ok()).toBeTruthy();
      const commentCreateBody = await createCommentRes.json();
      const commentID = Number(commentCreateBody.data?.id || 0);
      expect(commentID).toBeGreaterThan(0);
      const commentListAfterCreate = await request.get(`${APP_BASE_URL}/v1/admin/comments?page=1&page_size=100`, { headers });
      const commentListAfterCreateBody = await commentListAfterCreate.json();
      expect((commentListAfterCreateBody.data?.list || []).some((item) => item.id === commentID)).toBeTruthy();
      await page.locator('.sidebar-menu .menu-item').filter({ hasText: '互动管理' }).click();
      await page.locator('.submenu .menu-item').filter({ hasText: '评论管理' }).click();
      await expect(page.getByText(commentText).first()).toBeVisible({ timeout: 15000 });
      const commentAuditRes = await request.put(`${APP_BASE_URL}/v1/admin/comments/${commentID}/audit`, {
        headers,
        data: { status: 1 },
      });
      expect(commentAuditRes.ok()).toBeTruthy();
      const commentListAfterAudit = await request.get(`${APP_BASE_URL}/v1/admin/comments?page=1&page_size=100`, { headers });
      const commentListAfterAuditBody = await commentListAfterAudit.json();
      expect((commentListAfterAuditBody.data?.list || []).some((item) => item.id === commentID && Number(item.status) === 1)).toBeTruthy();
      const commentDeleteRes = await request.delete(`${APP_BASE_URL}/v1/admin/comments/${commentID}`, { headers });
      expect(commentDeleteRes.ok()).toBeTruthy();
      const commentListAfterDelete = await request.get(`${APP_BASE_URL}/v1/admin/comments?page=1&page_size=100`, { headers });
      const commentListAfterDeleteBody = await commentListAfterDelete.json();
      expect((commentListAfterDeleteBody.data?.list || []).some((item) => item.id === commentID)).toBeFalsy();

      // 7) course + units CRUD
      const courseCreateRes = await request.post(`${APP_BASE_URL}/v1/admin/courses`, {
        headers,
        data: { title: courseTitle, description: 'ui-crud-course', price: 0, module_id: moduleID, status: 0 },
      });
      expect(courseCreateRes.ok()).toBeTruthy();
      const courseCreateBody = await courseCreateRes.json();
      const courseID = Number(courseCreateBody.data?.id || 0);
      expect(courseID).toBeGreaterThan(0);
      const courseListAfterCreate = await request.get(
        `${APP_BASE_URL}/v1/admin/courses?page=1&page_size=100&keyword=${encodeURIComponent(courseTitle)}`,
        { headers },
      );
      const courseListAfterCreateBody = await courseListAfterCreate.json();
      expect((courseListAfterCreateBody.data?.list || []).some((item) => item.id === courseID)).toBeTruthy();
      await page.getByText('课程管理').click();
      await page.locator('.search-input').first().fill(courseTitle);
      await page.keyboard.press('Enter');
      await expect(page.getByText(courseTitle).first()).toBeVisible({ timeout: 15000 });
      const courseUpdatedTitle = `${courseTitle} Updated`;
      const courseUpdateRes = await request.put(`${APP_BASE_URL}/v1/admin/courses/${courseID}`, {
        headers,
        data: { title: courseUpdatedTitle, description: 'ui-crud-course-updated', price: 0, module_id: moduleID, status: 0 },
      });
      expect(courseUpdateRes.ok()).toBeTruthy();
      const courseListAfterUpdate = await request.get(
        `${APP_BASE_URL}/v1/admin/courses?page=1&page_size=100&keyword=${encodeURIComponent(courseUpdatedTitle)}`,
        { headers },
      );
      const courseListAfterUpdateBody = await courseListAfterUpdate.json();
      expect((courseListAfterUpdateBody.data?.list || []).some((item) => item.id === courseID)).toBeTruthy();

      const unitCreateRes = await request.post(`${APP_BASE_URL}/v1/admin/courses/${courseID}/units`, {
        headers,
        data: { title: unitTitle, video_file_id: 0, duration: UNIT_DURATION_INITIAL, sort_order: 1 },
      });
      expect(unitCreateRes.ok()).toBeTruthy();
      const unitCreateBody = await unitCreateRes.json();
      const unitID = Number(unitCreateBody.data?.id || 0);
      expect(unitID).toBeGreaterThan(0);
      const unitListAfterCreate = await request.get(`${APP_BASE_URL}/v1/admin/courses/${courseID}/units`, { headers });
      const unitListAfterCreateBody = await unitListAfterCreate.json();
      expect((unitListAfterCreateBody.data?.list || unitListAfterCreateBody.data || []).some((item) => item.id === unitID)).toBeTruthy();
      const unitUpdatedTitle = `${unitTitle} Updated`;
      const unitUpdateRes = await request.put(`${APP_BASE_URL}/v1/admin/courses/${courseID}/units/${unitID}`, {
        headers,
        data: { title: unitUpdatedTitle, video_file_id: 0, duration: UNIT_DURATION_UPDATED, sort_order: 2 },
      });
      expect(unitUpdateRes.ok()).toBeTruthy();
      const unitListAfterUpdate = await request.get(`${APP_BASE_URL}/v1/admin/courses/${courseID}/units`, { headers });
      const unitListAfterUpdateBody = await unitListAfterUpdate.json();
      expect((unitListAfterUpdateBody.data?.list || unitListAfterUpdateBody.data || []).some((item) => item.id === unitID && item.title === unitUpdatedTitle)).toBeTruthy();
      const unitDeleteRes = await request.delete(`${APP_BASE_URL}/v1/admin/courses/${courseID}/units/${unitID}`, { headers });
      expect(unitDeleteRes.ok()).toBeTruthy();
      const unitListAfterDelete = await request.get(`${APP_BASE_URL}/v1/admin/courses/${courseID}/units`, { headers });
      const unitListAfterDeleteBody = await unitListAfterDelete.json();
      expect((unitListAfterDeleteBody.data?.list || unitListAfterDeleteBody.data || []).some((item) => item.id === unitID)).toBeFalsy();
      const courseDeleteRes = await request.delete(`${APP_BASE_URL}/v1/admin/courses/${courseID}`, { headers });
      expect(courseDeleteRes.ok()).toBeTruthy();
      const courseListAfterDelete = await request.get(
        `${APP_BASE_URL}/v1/admin/courses?page=1&page_size=100&keyword=${encodeURIComponent(courseUpdatedTitle)}`,
        { headers },
      );
      const courseListAfterDeleteBody = await courseListAfterDelete.json();
      expect((courseListAfterDeleteBody.data?.list || []).some((item) => item.id === courseID)).toBeFalsy();

      // 8) banner CRUD
      const uploadName = `ui-crud-banner-${unique}.png`;
      const presignRes = await request.get(
        `${APP_BASE_URL}/v1/admin/upload/files/presign?filename=${encodeURIComponent(uploadName)}&usage=protected&expires_in=600`,
        { headers },
      );
      expect(presignRes.ok()).toBeTruthy();
      const presignBody = await presignRes.json();
      const bannerFileID = Number(presignBody.data?.file_id || 0);
      expect(bannerFileID).toBeGreaterThan(0);
      const putRes = await request.fetch(presignBody.data?.put_url || '', {
        method: 'PUT',
        headers: { 'Content-Type': 'image/png' },
        data: Buffer.from('ui-crud-banner-file'),
      });
      expect(putRes.ok()).toBeTruthy();
      const bannerCreateRes = await request.post(`${APP_BASE_URL}/v1/admin/banners`, {
        headers,
        data: { title: bannerTitle, image_file_id: bannerFileID, link_url: 'https://example.com/ui-crud', sort_order: 0, status: 1 },
      });
      expect(bannerCreateRes.ok()).toBeTruthy();
      const bannerCreateBody = await bannerCreateRes.json();
      const bannerID = Number(bannerCreateBody.data?.id || 0);
      expect(bannerID).toBeGreaterThan(0);
      const bannerListAfterCreate = await request.get(`${APP_BASE_URL}/v1/admin/banners`, { headers });
      const bannerListAfterCreateBody = await bannerListAfterCreate.json();
      expect((bannerListAfterCreateBody.data?.list || bannerListAfterCreateBody.data || []).some((item) => item.id === bannerID)).toBeTruthy();
      await page.getByText('轮播图管理').click();
      await expect(page.getByText(bannerTitle).first()).toBeVisible({ timeout: 15000 });
      const bannerUpdatedTitle = `${bannerTitle} Updated`;
      const bannerUpdateRes = await request.put(`${APP_BASE_URL}/v1/admin/banners/${bannerID}`, {
        headers,
        data: { title: bannerUpdatedTitle, image_file_id: bannerFileID, link_url: 'https://example.com/ui-crud-updated', sort_order: 1, status: 1 },
      });
      expect(bannerUpdateRes.ok()).toBeTruthy();
      const bannerListAfterUpdate = await request.get(`${APP_BASE_URL}/v1/admin/banners`, { headers });
      const bannerListAfterUpdateBody = await bannerListAfterUpdate.json();
      expect((bannerListAfterUpdateBody.data?.list || bannerListAfterUpdateBody.data || []).some((item) => item.id === bannerID && item.title === bannerUpdatedTitle)).toBeTruthy();
      const bannerDeleteRes = await request.delete(`${APP_BASE_URL}/v1/admin/banners/${bannerID}`, { headers });
      expect(bannerDeleteRes.ok()).toBeTruthy();
      const bannerListAfterDelete = await request.get(`${APP_BASE_URL}/v1/admin/banners`, { headers });
      const bannerListAfterDeleteBody = await bannerListAfterDelete.json();
      expect((bannerListAfterDeleteBody.data?.list || bannerListAfterDeleteBody.data || []).some((item) => item.id === bannerID)).toBeFalsy();

      // 9) finish article/module delete and verify delete query
      const articleDeleteRes = await request.delete(`${APP_BASE_URL}/v1/admin/articles/${articleID}`, { headers });
      expect(articleDeleteRes.ok()).toBeTruthy();
      const articleListAfterDelete = await request.get(
        `${APP_BASE_URL}/v1/admin/articles?page=1&page_size=100&keyword=${encodeURIComponent(articleUpdatedTitle)}`,
        { headers },
      );
      const articleListAfterDeleteBody = await articleListAfterDelete.json();
      expect((articleListAfterDeleteBody.data?.list || []).some((item) => item.id === articleID)).toBeFalsy();

      const moduleDeleteRes = await request.delete(`${APP_BASE_URL}/v1/admin/modules/${moduleID}`, { headers });
      expect(moduleDeleteRes.ok()).toBeTruthy();
      const moduleListAfterDelete = await request.get(
        `${APP_BASE_URL}/v1/modules?page=1&page_size=100&keyword=${encodeURIComponent(moduleUpdatedTitle)}`,
        { headers },
      );
      const moduleListAfterDeleteBody = await moduleListAfterDelete.json();
      expect((moduleListAfterDeleteBody.data?.list || moduleListAfterDeleteBody.data || []).some((item) => item.id === moduleID)).toBeFalsy();
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
      const moduleID = await ensureModuleIDForArticle(request, adminToken);

      const articleRes = await request.post(`${APP_BASE_URL}/v1/admin/articles`, {
        headers: { Authorization: `Bearer ${adminToken}` },
        data: { title: `UI Notif ${Date.now()}`, summary: 's', content: 'c', content_type: 1, module_id: moduleID },
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

    test('miniprogram user interactions for articles and courses', async ({ page, request }, testInfo) => {
      const adminToken = await getAdminToken(request);
      const userToken = await getUserToken(request);
      const headers = { Authorization: `Bearer ${adminToken}` };
      const unique = `${Date.now()}-${testInfo.retry}-${testInfo.workerIndex}`;
      const moduleID = await ensureModuleIDForArticle(request, adminToken);

      const articleTitle = `UI MP CRUD Article ${unique}`;
      const articleComment = `ui-mp-article-comment-${unique}`;
      const courseTitle = `UI MP CRUD Course ${unique}`;
      const courseComment = `ui-mp-course-comment-${unique}`;

      const articleRes = await request.post(`${APP_BASE_URL}/v1/admin/articles`, {
        headers,
        data: { title: articleTitle, summary: 'ui-mp-crud', content: 'ui-mp-crud-content', content_type: 1, module_id: moduleID },
      });
      expect(articleRes.ok()).toBeTruthy();
      const articleBody = await articleRes.json();
      const articleID = Number(articleBody.data?.id || 0);
      expect(articleID).toBeGreaterThan(0);
      const articlePublishRes = await request.post(`${APP_BASE_URL}/v1/admin/articles/${articleID}/publish`, {
        headers,
        data: { status: 1 },
      });
      expect(articlePublishRes.ok()).toBeTruthy();
      const publicArticleQuery = await request.get(`${APP_BASE_URL}/v1/articles?page=1&page_size=20&keyword=${encodeURIComponent(articleTitle)}`, {
        headers: { Authorization: `Bearer ${userToken}` },
      });
      const publicArticleQueryBody = await publicArticleQuery.json();
      expect((publicArticleQueryBody.data?.list || []).some((item) => item.id === articleID)).toBeTruthy();

      const courseRes = await request.post(`${APP_BASE_URL}/v1/admin/courses`, {
        headers,
        data: { title: courseTitle, description: 'ui-mp-crud-course', price: 0, module_id: moduleID, status: 0 },
      });
      expect(courseRes.ok()).toBeTruthy();
      const courseBody = await courseRes.json();
      const courseID = Number(courseBody.data?.id || 0);
      expect(courseID).toBeGreaterThan(0);
      const coursePublishRes = await request.post(`${APP_BASE_URL}/v1/admin/courses/${courseID}/publish`, {
        headers,
        data: { status: 1 },
      });
      expect(coursePublishRes.ok()).toBeTruthy();
      const publicCourseQuery = await request.get(`${APP_BASE_URL}/v1/courses?page=1&page_size=20&keyword=${encodeURIComponent(courseTitle)}`, {
        headers: { Authorization: `Bearer ${userToken}` },
      });
      const publicCourseQueryBody = await publicCourseQuery.json();
      expect((publicCourseQueryBody.data?.list || []).some((item) => item.id === courseID)).toBeTruthy();

      // article create/read/delete by UI (like, collection, comment)
      await page.locator('.tab-item').filter({ hasText: '文章' }).click();
      await page.locator('input[placeholder*="搜索文章"]').fill(articleTitle);
      const articleCardTitle = page.locator('.card:visible .card-title').filter({ hasText: articleTitle }).first();
      await expect(articleCardTitle).toBeVisible({ timeout: 15000 });
      await articleCardTitle.click();
      await expect(page.locator('.detail-title').filter({ hasText: '文章详情' })).toBeVisible({ timeout: 15000 });
      await expect(page.getByText(articleTitle).first()).toBeVisible();
      await page.locator('.detail-actions .action-btn').first().click();
      await expect(page.locator('.detail-actions .action-btn.liked')).toBeVisible();
      await page.locator('.detail-actions .action-btn').nth(1).click();
      await expect(page.locator('.detail-actions .action-btn').nth(1)).toContainText('已收藏');
      await page.locator('textarea[placeholder="写评论..."]').fill(articleComment);
      await page.getByRole('button', { name: '发送' }).click();
      await expect(page.getByText(articleComment).first()).toBeVisible({ timeout: 15000 });
      const commentListRes = await request.get(`${APP_BASE_URL}/v1/comments/1/${articleID}`, {
        headers: { Authorization: `Bearer ${userToken}` },
      });
      const commentListBody = await commentListRes.json();
      expect((commentListBody.data?.list || commentListBody.data || []).some((item) => item.content === articleComment)).toBeTruthy();
      await page.locator('.back-btn').click();
      await page.locator('.tab-item').filter({ hasText: '我的' }).click();
      await page.getByText('我的收藏').click();
      await expect(
        page.locator('.sub-page .card-title').filter({ hasText: new RegExp(`${escapeRegExp(articleTitle)}|收藏\\s*#${articleID}`) }).first(),
      ).toBeVisible({ timeout: 15000 });
      await page.locator('.detail-header .back-btn').first().click();

      await page.locator('.tab-item').filter({ hasText: '文章' }).click();
      await page.locator('input[placeholder*="搜索文章"]').fill(articleTitle);
      const articleCardTitleAgain = page.locator('.card:visible .card-title').filter({ hasText: articleTitle }).first();
      await expect(articleCardTitleAgain).toBeVisible({ timeout: 15000 });
      await articleCardTitleAgain.click();
      await page.locator('.detail-actions .action-btn').nth(1).click();
      await expect(page.locator('.detail-actions .action-btn').nth(1)).toContainText('收藏');
      await page.locator('.detail-actions .action-btn').first().click();
      await page.locator('.back-btn').click();

      // course create/read/delete by UI (like, collection, comment)
      await page.locator('.tab-item').filter({ hasText: '课程' }).click();
      await page.locator('input[placeholder*="搜索课程"]').fill(courseTitle);
      const courseCardTitle = page.locator('.card:visible .card-title').filter({ hasText: courseTitle }).first();
      await expect(courseCardTitle).toBeVisible({ timeout: 15000 });
      await courseCardTitle.click();
      await expect(page.locator('.detail-title').filter({ hasText: '课程详情' })).toBeVisible({ timeout: 15000 });
      await expect(page.getByText(courseTitle).first()).toBeVisible();
      await page.locator('.detail-actions .action-btn').first().click();
      await expect(page.locator('.detail-actions .action-btn.liked')).toBeVisible();
      await page.locator('.detail-actions .action-btn').nth(1).click();
      await expect(page.locator('.detail-actions .action-btn').nth(1)).toContainText('已收藏');
      await page.locator('textarea[placeholder="写评论..."]').fill(courseComment);
      await page.getByRole('button', { name: '发送' }).click();
      await expect(page.getByText(courseComment).first()).toBeVisible({ timeout: 15000 });
      const courseCommentListRes = await request.get(`${APP_BASE_URL}/v1/comments/2/${courseID}`, {
        headers: { Authorization: `Bearer ${userToken}` },
      });
      const courseCommentListBody = await courseCommentListRes.json();
      expect((courseCommentListBody.data?.list || courseCommentListBody.data || []).some((item) => item.content === courseComment)).toBeTruthy();
      await page.locator('.back-btn').click();

      await page.locator('.tab-item').filter({ hasText: '课程' }).click();
      await page.locator('input[placeholder*="搜索课程"]').fill(courseTitle);
      const courseCardTitleAgain = page.locator('.card:visible .card-title').filter({ hasText: courseTitle }).first();
      await expect(courseCardTitleAgain).toBeVisible({ timeout: 15000 });
      await courseCardTitleAgain.click();
      await page.locator('.detail-actions .action-btn').nth(1).click();
      await expect(page.locator('.detail-actions .action-btn').nth(1)).toContainText('收藏');
      await page.locator('.detail-actions .action-btn').first().click();

      // backend query confirms delete states applied by UI
      const articleDetailRes = await request.get(`${APP_BASE_URL}/v1/articles/${articleID}`, {
        headers: { Authorization: `Bearer ${userToken}` },
      });
      const articleDetailBody = await articleDetailRes.json();
      expect(articleDetailBody.data?.is_liked).toBeFalsy();
      expect(articleDetailBody.data?.is_collected).toBeFalsy();
      const courseDetailRes = await request.get(`${APP_BASE_URL}/v1/courses/${courseID}`, {
        headers: { Authorization: `Bearer ${userToken}` },
      });
      const courseDetailBody = await courseDetailRes.json();
      expect(courseDetailBody.data?.is_liked).toBeFalsy();
      expect(courseDetailBody.data?.is_collected).toBeFalsy();
    });
  });
});
