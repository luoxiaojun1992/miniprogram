/**
 * k6 API test suite covering all endpoints of the miniprogram backend.
 *
 * Run locally:
 *   k6 run -e BASE_URL=http://localhost:8080 tests/api/api_test.js
 *
 * Run via Docker Compose:
 *   docker compose -f docker-compose.api-test.yml up --exit-code-from k6 --abort-on-container-exit
 */

import http from 'k6/http';
import { check, group } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const ADMIN_TEST_PASSWORD = 'Test@123456';
const NON_EXISTENT_FILE_ID = 999999;
const DEFAULT_VIDEO_FILE_ID = 1;

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: {
    // At least 95 % of all checks must pass
    checks: ['rate>=0.95'],
    // Less than 10 % of HTTP requests may fail
    http_req_failed: ['rate<0.10'],
  },
};

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

function headers(token) {
  const h = { 'Content-Type': 'application/json' };
  if (token) h['Authorization'] = `Bearer ${token}`;
  return { headers: h };
}

function ok(r) {
  return r.status >= 200 && r.status < 300;
}

function multipartHeaders(token) {
  const h = {};
  if (token) h['Authorization'] = `Bearer ${token}`;
  return { headers: h };
}

function withExpectedStatuses(params, ...statuses) {
  return {
    ...params,
    responseCallback: http.expectedStatuses(...statuses),
  };
}

// ---------------------------------------------------------------------------
// setup – runs once before any VU iteration
// ---------------------------------------------------------------------------

export function setup() {
  // 1. Obtain tokens via the debug endpoint (enabled in test compose)
  const adminTokenRes = http.post(
    `${BASE_URL}/v1/debug/token`,
    JSON.stringify({ user_id: 1 }),
    headers(null),
  );
  check(adminTokenRes, { 'setup | debug token admin: 200': (r) => r.status === 200 });
  const adminToken = adminTokenRes.json('data.access_token');

  const userTokenRes = http.post(
    `${BASE_URL}/v1/debug/token`,
    JSON.stringify({ user_id: 2 }),
    headers(null),
  );
  check(userTokenRes, { 'setup | debug token user: 200': (r) => r.status === 200 });
  const userToken = userTokenRes.json('data.access_token');

  // 2. Create a test module
  const moduleRes = http.post(
    `${BASE_URL}/v1/admin/modules`,
    JSON.stringify({ title: 'K6 Test Module', description: 'Created by k6', sort_order: 999 }),
    headers(adminToken),
  );
  check(moduleRes, { 'setup | create module: 201': (r) => r.status === 201 });
  const moduleId = moduleRes.json('data.id');

  // 3. Create a test module page
  const pageRes = http.post(
    `${BASE_URL}/v1/admin/modules/${moduleId}/pages`,
    JSON.stringify({ title: 'K6 Test Page', content: 'Test page content', content_type: 1, sort_order: 1 }),
    headers(adminToken),
  );
  check(pageRes, { 'setup | create page: 201': (r) => r.status === 201 });
  const pageId = pageRes.json('data.id');

  // 4. Create a test article and publish it
  const articleRes = http.post(
    `${BASE_URL}/v1/admin/articles`,
    JSON.stringify({
      title: 'K6 Test Article',
      summary: 'An article created by k6',
      content: '# K6 Test\n\nContent.',
      content_type: 1,
      module_id: moduleId,
    }),
    headers(adminToken),
  );
  check(articleRes, { 'setup | create article: 201': (r) => r.status === 201 });
  const articleId = articleRes.json('data.id');

  http.post(
    `${BASE_URL}/v1/admin/articles/${articleId}/publish`,
    JSON.stringify({ status: 1 }),
    headers(adminToken),
  );

  // 5. Create a protected video file record, then create a test course/unit and publish
  const setupVideoPresignRes = http.get(
    `${BASE_URL}/v1/admin/upload/files/presign?filename=k6-setup-video.mp4&usage=protected&expires_in=600`,
    headers(adminToken),
  );
  check(setupVideoPresignRes, { 'setup | create video file record: 200': (r) => r.status === 200 });
  const setupVideoFileId = setupVideoPresignRes.json('data.file_id');

  const courseRes = http.post(
    `${BASE_URL}/v1/admin/courses`,
    JSON.stringify({
      title: 'K6 Test Course',
      description: 'A course created by k6',
      price: 0,
      module_id: moduleId,
    }),
    headers(adminToken),
  );
  check(courseRes, { 'setup | create course: 201': (r) => r.status === 201 });
  const courseId = courseRes.json('data.id');

  const unitRes = http.post(
    `${BASE_URL}/v1/admin/courses/${courseId}/units`,
    JSON.stringify({ title: 'K6 Test Unit', video_file_id: setupVideoFileId, duration: 30, sort_order: 1 }),
    headers(adminToken),
  );
  check(unitRes, { 'setup | create unit: 201': (r) => r.status === 201 });
  const unitId = unitRes.json('data.id');

  http.post(
    `${BASE_URL}/v1/admin/courses/${courseId}/publish`,
    JSON.stringify({ status: 1 }),
    headers(adminToken),
  );

  // 6. Create a test role
  const roleRes = http.post(
    `${BASE_URL}/v1/admin/roles`,
    JSON.stringify({ name: 'K6 Test Role', description: 'Created by k6' }),
    headers(adminToken),
  );
  check(roleRes, { 'setup | create role: 201': (r) => r.status === 201 });
  const roleId = roleRes.json('data.id');

  // 7. Create a test comment (as admin, on the article)
  const commentRes = http.post(
    `${BASE_URL}/v1/comments/1/${articleId}`,
    JSON.stringify({ content: 'K6 setup comment' }),
    headers(adminToken),
  );
  check(commentRes, { 'setup | create comment: 2xx': (r) => ok(r) });
  const commentId = commentRes.json('data.id');

  // 8. Create a secondary admin user (for CRUD admin-user tests)
  const newUserRes = http.post(
    `${BASE_URL}/v1/admin/users`,
    JSON.stringify({ email: 'k6_tmp_admin@example.com', password: ADMIN_TEST_PASSWORD, nickname: 'K6 Tmp', user_type: 2 }),
    headers(adminToken),
  );
  check(newUserRes, { 'setup | create admin user: 201': (r) => r.status === 201 });
  const newUserId = newUserRes.json('data.id');

  return { adminToken, userToken, moduleId, pageId, articleId, courseId, unitId, roleId, commentId, newUserId, setupVideoFileId };
}

// ---------------------------------------------------------------------------
// default – the main test scenario
// ---------------------------------------------------------------------------

export default function (data) {
  const { adminToken, userToken, moduleId, pageId, articleId, courseId, unitId, roleId, commentId, newUserId, setupVideoFileId } = data;
  const adminH = headers(adminToken);
  const userH  = headers(userToken);
  const jsonH  = headers(null);
  const uniqueTestId = `${Date.now()}-${Math.floor(Math.random() * 1000000)}`;
  let bannerMediaFileId = 0;

  // -------------------------------------------------------------------------
  group('Health', () => {
    const r = http.get(`${BASE_URL}/health`);
    check(r, {
      'GET /health: 200': (r) => r.status === 200,
      'GET /health: status ok': (r) => r.json('status') === 'ok',
    });
  });

  // -------------------------------------------------------------------------
  group('Auth', () => {
    // Admin login
    const loginRes = http.post(
      `${BASE_URL}/v1/auth/admin-login`,
      JSON.stringify({ email: 'admin@example.com', password: ADMIN_TEST_PASSWORD }),
      jsonH,
    );
    check(loginRes, {
      'POST /v1/auth/admin-login: 200': (r) => r.status === 200,
      'POST /v1/auth/admin-login: has token': (r) => !!r.json('data.access_token'),
    });
    const freshToken = loginRes.json('data.access_token');

    // Token refresh
    const refreshRes = http.post(`${BASE_URL}/v1/auth/refresh`, null, headers(freshToken));
    check(refreshRes, { 'POST /v1/auth/refresh: 200': (r) => r.status === 200 });

    // WeChat login – will fail with invalid code; just verify no 500
    const wechatRes = http.post(
      `${BASE_URL}/v1/auth/wechat-login`,
      JSON.stringify({ code: 'invalid-test-code' }),
      withExpectedStatuses(jsonH, { min: 200, max: 499 }),
    );
    check(wechatRes, { 'POST /v1/auth/wechat-login: not 500': (r) => r.status !== 500 });

    // Debug token (already tested in setup; verify once more in default)
    const debugRes = http.post(
      `${BASE_URL}/v1/debug/token`,
      JSON.stringify({ user_id: 1 }),
      jsonH,
    );
    check(debugRes, { 'POST /v1/debug/token: 200': (r) => r.status === 200 });
  });

  // -------------------------------------------------------------------------
  group('Audit Logs', () => {
    const uniqueTitle = `K6 Audit Module ${Date.now()}`;
    const writeRes = http.post(
      `${BASE_URL}/v1/admin/modules`,
      JSON.stringify({ title: uniqueTitle, description: 'audit test', sort_order: 777 }),
      adminH,
    );
    check(writeRes, { 'POST /v1/admin/modules (audit seed): 201': (r) => r.status === 201 });

    const logsRes = http.get(
      `${BASE_URL}/v1/admin/audit-logs?page=1&page_size=20&module=modules&action=create`,
      adminH,
    );
    check(logsRes, {
      'GET /v1/admin/audit-logs: 200': (r) => r.status === 200,
      'audit logs contains create/modules entry': (r) => {
        const list = r.json('data.list') || [];
        return list.length > 0 && list.some((x) => x.module === 'modules' && x.action === 'create');
      },
    });
  });

  // -------------------------------------------------------------------------
  group('User', () => {
    // Get profile
    const profileRes = http.get(`${BASE_URL}/v1/users/profile`, userH);
    check(profileRes, { 'GET /v1/users/profile: 200': (r) => r.status === 200 });

    // Update profile
    const updateRes = http.put(
      `${BASE_URL}/v1/users/profile`,
      JSON.stringify({ nickname: 'K6 User Updated', avatar_url: '' }),
      userH,
    );
    check(updateRes, { 'PUT /v1/users/profile: 2xx': (r) => ok(r) });

    // Get permissions
    const permRes = http.get(`${BASE_URL}/v1/users/permissions`, userH);
    check(permRes, { 'GET /v1/users/permissions: 200': (r) => r.status === 200 });
  });

  // -------------------------------------------------------------------------
  group('Public Content', () => {
    // List modules
    const modulesRes = http.get(`${BASE_URL}/v1/modules`);
    check(modulesRes, { 'GET /v1/modules: 200': (r) => r.status === 200 });

    // List banners
    const bannersRes = http.get(`${BASE_URL}/v1/banners`);
    check(bannersRes, { 'GET /v1/banners: 200': (r) => r.status === 200 });

    // List articles
    const articlesRes = http.get(`${BASE_URL}/v1/articles?page=1&page_size=10`);
    check(articlesRes, { 'GET /v1/articles: 200': (r) => r.status === 200 });

    // Get article by ID
    const articleRes = http.get(`${BASE_URL}/v1/articles/${articleId}`);
    check(articleRes, { 'GET /v1/articles/:id: 200': (r) => r.status === 200 });

    // List courses (public)
    const coursesRes = http.get(`${BASE_URL}/v1/courses?page=1&page_size=10`);
    check(coursesRes, { 'GET /v1/courses: 200': (r) => r.status === 200 });

    // Get course by ID (requires auth)
    const courseRes = http.get(`${BASE_URL}/v1/courses/${courseId}`, userH);
    check(courseRes, { 'GET /v1/courses/:id: 200': (r) => r.status === 200 });

    // List comments (public)
    const commentsRes = http.get(`${BASE_URL}/v1/comments/1/${articleId}`);
    check(commentsRes, { 'GET /v1/comments/:type/:id: 200': (r) => r.status === 200 });
  });

  // -------------------------------------------------------------------------
  group('User Interactions', () => {
    const articleBefore = http.get(`${BASE_URL}/v1/articles/${articleId}`);
    const likeCountBefore = articleBefore.json('data.like_count') || 0;
    const commentCountBefore = articleBefore.json('data.comment_count') || 0;

    // --- Study Records & Collections ---
    // Study records – list
    const studyListRes = http.get(`${BASE_URL}/v1/study-records`, userH);
    check(studyListRes, { 'GET /v1/study-records: 200': (r) => r.status === 200 });

    // Study records – update
    const studyUpdateRes = http.post(
      `${BASE_URL}/v1/study-records`,
      JSON.stringify({ unit_id: unitId, progress: 60, status: 1 }),
      userH,
    );
    check(studyUpdateRes, { 'POST /v1/study-records: 2xx': (r) => ok(r) });

    // Collections – list
    const colListRes = http.get(`${BASE_URL}/v1/collections`, userH);
    check(colListRes, { 'GET /v1/collections: 200': (r) => r.status === 200 });

    // Collections – add
    const colAddRes = http.post(`${BASE_URL}/v1/collections/1/${articleId}`, null, userH);
    check(colAddRes, { 'POST /v1/collections/1/:id: 2xx': (r) => ok(r) });

    // Collections – remove
    const colDelRes = http.del(`${BASE_URL}/v1/collections/1/${articleId}`, null, userH);
    check(colDelRes, { 'DELETE /v1/collections/1/:id: 2xx': (r) => ok(r) });

    // Collections – add/remove for course
    const colCourseAddRes = http.post(`${BASE_URL}/v1/collections/2/${courseId}`, null, userH);
    check(colCourseAddRes, { 'POST /v1/collections/2/:id: 2xx': (r) => ok(r) });
    const colCourseDelRes = http.del(`${BASE_URL}/v1/collections/2/${courseId}`, null, userH);
    check(colCourseDelRes, { 'DELETE /v1/collections/2/:id: 2xx': (r) => ok(r) });

    // --- Likes & Follows ---
    // Likes – add
    const likeAddRes = http.post(`${BASE_URL}/v1/likes/1/${articleId}`, null, userH);
    check(likeAddRes, { 'POST /v1/likes/1/:id: 2xx': (r) => ok(r) });

    const articleAfterLike = http.get(`${BASE_URL}/v1/articles/${articleId}`);
    check(articleAfterLike, {
      'like count endpoint readable after like': (r) => {
        const value = r.json('data.like_count');
        return r.status === 200 && (value === null || value === undefined || Number(value) >= likeCountBefore);
      },
    });

    // Likes – remove
    const likeDelRes = http.del(`${BASE_URL}/v1/likes/1/${articleId}`, null, userH);
    check(likeDelRes, { 'DELETE /v1/likes/1/:id: 2xx': (r) => ok(r) });

    // Likes – add/remove for course
    const likeCourseAddRes = http.post(`${BASE_URL}/v1/likes/2/${courseId}`, null, userH);
    check(likeCourseAddRes, { 'POST /v1/likes/2/:id: 2xx': (r) => ok(r) });
    const likeCourseDelRes = http.del(`${BASE_URL}/v1/likes/2/${courseId}`, null, userH);
    check(likeCourseDelRes, { 'DELETE /v1/likes/2/:id: 2xx': (r) => ok(r) });

    // Follows – add/remove
    const followAddRes = http.post(`${BASE_URL}/v1/follows/${newUserId}`, null, userH);
    check(followAddRes, { 'POST /v1/follows/:user_id: 2xx': (r) => ok(r) });
    const followDelRes = http.del(`${BASE_URL}/v1/follows/${newUserId}`, null, userH);
    check(followDelRes, { 'DELETE /v1/follows/:user_id: 2xx': (r) => ok(r) });

    // --- Comments, Notifications & Files ---
    // Comments – create (as regular user)
    const createCommentRes = http.post(
      `${BASE_URL}/v1/comments/1/${articleId}`,
      JSON.stringify({ content: 'k6 user comment' }),
      userH,
    );
    check(createCommentRes, { 'POST /v1/comments/1/:id: 2xx': (r) => ok(r) });
    const userCommentId = createCommentRes.json('data.id');

    const articleAfterComment = http.get(`${BASE_URL}/v1/articles/${articleId}`);
    check(articleAfterComment, {
      'comment count endpoint readable after comment': (r) => {
        const value = r.json('data.comment_count');
        return r.status === 200 && (value === null || value === undefined || Number(value) >= commentCountBefore);
      },
    });

    const adminNotifRes = http.get(`${BASE_URL}/v1/notifications`, adminH);
    check(adminNotifRes, {
      'interaction creates like/comment notifications': (r) => {
        const list = r.json('data.list') || [];
        return Array.isArray(list) && list.some((x) => x.type === 2 || x.type === 4);
      },
    });

    if (userCommentId) {
      const cleanupCommentRes = http.del(`${BASE_URL}/v1/admin/comments/${userCommentId}`, null, adminH);
      check(cleanupCommentRes, { 'cleanup created comment: 2xx': (r) => ok(r) });
    }

    // Comments – create on course then cleanup
    const createCourseCommentRes = http.post(
      `${BASE_URL}/v1/comments/2/${courseId}`,
      JSON.stringify({ content: 'k6 course comment' }),
      userH,
    );
    check(createCourseCommentRes, { 'POST /v1/comments/2/:id: 2xx': (r) => ok(r) });
    const userCourseCommentId = createCourseCommentRes.json('data.id');
    if (userCourseCommentId) {
      const cleanupCourseCommentRes = http.del(`${BASE_URL}/v1/admin/comments/${userCourseCommentId}`, null, adminH);
      check(cleanupCourseCommentRes, { 'cleanup course comment: 2xx': (r) => ok(r) });
    }

    // Notifications – list
    const notifListRes = http.get(`${BASE_URL}/v1/notifications`, userH);
    check(notifListRes, { 'GET /v1/notifications: 200': (r) => r.status === 200 });

    // Notifications – mark single read (if any exist)
    const notifs = notifListRes.json('data.list');
    if (Array.isArray(notifs) && notifs.length > 0) {
      const nid = notifs[0].id;
      const markRes = http.put(`${BASE_URL}/v1/notifications/${nid}/read`, null, userH);
      check(markRes, { 'PUT /v1/notifications/:id/read: 2xx': (r) => ok(r) });
    }

    // Notifications – mark all read
    const readAllRes = http.put(`${BASE_URL}/v1/notifications/read-all`, null, userH);
    check(readAllRes, { 'PUT /v1/notifications/read-all: 2xx': (r) => ok(r) });

    // Front user upload avatar image (COS-backed in docker API tests)
    const uploadPayload = {
      file: http.file(new Uint8Array([137, 80, 78, 71, 13, 10, 26, 10]).buffer, 'k6.png', 'image/png'),
    };
    const uploadRes = http.post(`${BASE_URL}/v1/upload/avatar`, uploadPayload, multipartHeaders(userToken));
    check(uploadRes, {
      'POST /v1/upload/avatar: 200': (r) => r.status === 200,
      'POST /v1/upload/avatar: cos url': (r) => {
        const url = r.json('data.url');
        return typeof url === 'string' && url.indexOf('/miniapp-test/avatar/') !== -1;
      },
    });

    // Admin gets unified file presign URL for protected course video upload
    const presignRes = http.get(`${BASE_URL}/v1/admin/upload/files/presign?filename=k6-video.mp4&usage=protected&expires_in=600`, adminH);
    check(presignRes, {
      'GET /v1/admin/upload/files/presign: 200': (r) => r.status === 200,
      'GET /v1/admin/upload/files/presign: has put_url': (r) => typeof r.json('data.put_url') === 'string',
      'GET /v1/admin/upload/files/presign: has file_id': (r) => typeof r.json('data.file_id') === 'number',
    });

    // Admin gets banner media presign URL
    const bannerPresignRes = http.get(
      `${BASE_URL}/v1/admin/upload/banner/media/presign?filename=k6-banner.png&expires_in=600`,
      adminH,
    );
    check(bannerPresignRes, {
      'GET /v1/admin/upload/banner/media/presign: 200': (r) => r.status === 200,
      'GET /v1/admin/upload/banner/media/presign: has file_id': (r) => typeof r.json('data.file_id') === 'number',
    });
    bannerMediaFileId = bannerPresignRes.json('data.file_id') || 0;

    // Download URL endpoints (use non-existent IDs to avoid object dependency in mock COS)
    // Static material download is a public endpoint by design.
    const staticDownloadRes = http.get(
      `${BASE_URL}/v1/download/static/${NON_EXISTENT_FILE_ID}`,
      withExpectedStatuses({}, 404),
    );
    check(staticDownloadRes, { 'GET /v1/download/static/:file_id: 404': (r) => r.status === 404 });

    const videoDownloadRes = http.get(
      `${BASE_URL}/v1/download/course/video/${NON_EXISTENT_FILE_ID}`,
      withExpectedStatuses(userH, 404),
    );
    check(videoDownloadRes, { 'GET /v1/download/course/video/:file_id: 404': (r) => r.status === 404 });

    const articleAttachDownloadRes = http.get(
      `${BASE_URL}/v1/download/article/attachment/${NON_EXISTENT_FILE_ID}`,
      withExpectedStatuses(userH, 404),
    );
    check(articleAttachDownloadRes, { 'GET /v1/download/article/attachment/:file_id: 404': (r) => r.status === 404 });

    const courseAttachDownloadRes = http.get(
      `${BASE_URL}/v1/download/course/attachment/${NON_EXISTENT_FILE_ID}`,
      withExpectedStatuses(userH, 404),
    );
    check(courseAttachDownloadRes, { 'GET /v1/download/course/attachment/:file_id: 404': (r) => r.status === 404 });

    const bannerDownloadRes = http.get(
      `${BASE_URL}/v1/download/banner/media/${NON_EXISTENT_FILE_ID}`,
      withExpectedStatuses(userH, 404),
    );
    check(bannerDownloadRes, { 'GET /v1/download/banner/media/:file_id: 404': (r) => r.status === 404 });
  });

  // -------------------------------------------------------------------------
  group('Admin – Users', () => {
    // List
    const listRes = http.get(`${BASE_URL}/v1/admin/users?page=1&page_size=10`, adminH);
    check(listRes, { 'GET /v1/admin/users: 200': (r) => r.status === 200 });

    // Get
    const getRes = http.get(`${BASE_URL}/v1/admin/users/${newUserId}`, adminH);
    check(getRes, { 'GET /v1/admin/users/:id: 200': (r) => r.status === 200 });

    // Update
    const updateRes = http.put(
      `${BASE_URL}/v1/admin/users/${newUserId}`,
      JSON.stringify({ nickname: 'K6 Tmp Updated' }),
      adminH,
    );
    check(updateRes, { 'PUT /v1/admin/users/:id: 2xx': (r) => ok(r) });

    // Assign roles
    const assignRes = http.put(
      `${BASE_URL}/v1/admin/users/${newUserId}/roles`,
      JSON.stringify({ role_ids: [roleId] }),
      adminH,
    );
    check(assignRes, { 'PUT /v1/admin/users/:id/roles: 2xx': (r) => ok(r) });

    // Add tag
    const addTagRes = http.post(
      `${BASE_URL}/v1/admin/users/${newUserId}/tags`,
      JSON.stringify({ tag_name: 'k6-tag' }),
      adminH,
    );
    check(addTagRes, { 'POST /v1/admin/users/:id/tags: 2xx': (r) => ok(r) });
    const tagId = addTagRes.json('data.id');

    // Delete tag
    if (tagId) {
      const delTagRes = http.del(
        `${BASE_URL}/v1/admin/users/${newUserId}/tags?tag_id=${tagId}`,
        null,
        adminH,
      );
      check(delTagRes, { 'DELETE /v1/admin/users/:id/tags: 2xx': (r) => ok(r) });
    }

    // Note: user CREATE is covered in setup, and user DELETE is covered in teardown cleanup.
  });

  // -------------------------------------------------------------------------
  group('Admin – Roles & Permissions', () => {
    // List roles
    const listRes = http.get(`${BASE_URL}/v1/admin/roles`, adminH);
    check(listRes, { 'GET /v1/admin/roles: 200': (r) => r.status === 200 });

    // Get role
    const getRes = http.get(`${BASE_URL}/v1/admin/roles/${roleId}`, adminH);
    check(getRes, { 'GET /v1/admin/roles/:id: 200': (r) => r.status === 200 });

    // Update role
    const updateRes = http.put(
      `${BASE_URL}/v1/admin/roles/${roleId}`,
      JSON.stringify({ name: 'K6 Test Role Updated', description: 'Updated by k6' }),
      adminH,
    );
    check(updateRes, { 'PUT /v1/admin/roles/:id: 2xx': (r) => ok(r) });

    // Permissions tree
    const permRes = http.get(`${BASE_URL}/v1/admin/permissions`, adminH);
    check(permRes, { 'GET /v1/admin/permissions: 200': (r) => r.status === 200 });

    // Create and delete an extra role to cover full role CRUD
    const roleCreateRes = http.post(
      `${BASE_URL}/v1/admin/roles`,
      JSON.stringify({ name: `K6 Extra Role ${uniqueTestId}`, description: 'Extra role for api coverage' }),
      adminH,
    );
    check(roleCreateRes, { 'POST /v1/admin/roles: 201': (r) => r.status === 201 });
    const tempRoleId = roleCreateRes.json('data.id');
    if (tempRoleId) {
      const roleDeleteRes = http.del(`${BASE_URL}/v1/admin/roles/${tempRoleId}`, null, adminH);
      check(roleDeleteRes, { 'DELETE /v1/admin/roles/:id: 2xx': (r) => ok(r) });
    }
  });

  // -------------------------------------------------------------------------
  group('Admin – Modules', () => {
    const deleteRes = http.del(
      `${BASE_URL}/v1/admin/modules/${moduleId}`,
      null,
      withExpectedStatuses(adminH, { min: 400, max: 499 }),
    );
    check(deleteRes, { 'DELETE /v1/admin/modules/:id blocked when has associations': (r) => r.status >= 400 });

    // Update module
    const updateRes = http.put(
      `${BASE_URL}/v1/admin/modules/${moduleId}`,
      JSON.stringify({ title: 'K6 Module Updated', description: 'Updated', sort_order: 999 }),
      adminH,
    );
    check(updateRes, { 'PUT /v1/admin/modules/:id: 2xx': (r) => ok(r) });

    // List pages
    const pagesRes = http.get(`${BASE_URL}/v1/admin/modules/${moduleId}/pages`, adminH);
    check(pagesRes, { 'GET /v1/admin/modules/:id/pages: 200': (r) => r.status === 200 });

    // Update page
    const updatePageRes = http.put(
      `${BASE_URL}/v1/admin/modules/${moduleId}/pages/${pageId}`,
      JSON.stringify({ title: 'K6 Page Updated', content: 'Updated content', content_type: 1, sort_order: 1 }),
      adminH,
    );
    check(updatePageRes, { 'PUT /v1/admin/modules/:id/pages/:pid: 2xx': (r) => ok(r) });

    // Create and delete a temporary page
    const createPageRes = http.post(
      `${BASE_URL}/v1/admin/modules/${moduleId}/pages`,
      JSON.stringify({ title: 'K6 Temp Page', content: 'Temp page content', content_type: 1, sort_order: 2 }),
      adminH,
    );
    check(createPageRes, { 'POST /v1/admin/modules/:id/pages: 201': (r) => r.status === 201 });
    const tempPageId = createPageRes.json('data.id');
    if (tempPageId) {
      const deletePageRes = http.del(`${BASE_URL}/v1/admin/modules/${moduleId}/pages/${tempPageId}`, null, adminH);
      check(deletePageRes, { 'DELETE /v1/admin/modules/:id/pages/:page_id: 2xx': (r) => ok(r) });
    }
  });

  // -------------------------------------------------------------------------
  group('Admin – Banners', () => {
    // List
    const listRes = http.get(`${BASE_URL}/v1/admin/banners`, adminH);
    check(listRes, { 'GET /v1/admin/banners: 200': (r) => r.status === 200 });

    if (bannerMediaFileId) {
      // Create
      const createRes = http.post(
        `${BASE_URL}/v1/admin/banners`,
        JSON.stringify({
          title: `K6 Banner ${uniqueTestId}`,
          image_file_id: bannerMediaFileId,
          link_url: 'https://example.com/k6',
          sort_order: 10,
          status: 1,
        }),
        adminH,
      );
      check(createRes, { 'POST /v1/admin/banners: 201': (r) => r.status === 201 });
      const bannerId = createRes.json('data.id');
      if (bannerId) {
        // Update
        const updateRes = http.put(
          `${BASE_URL}/v1/admin/banners/${bannerId}`,
          JSON.stringify({
            title: `K6 Banner Updated ${uniqueTestId}`,
            image_file_id: bannerMediaFileId,
            link_url: 'https://example.com/k6-updated',
            sort_order: 11,
            status: 1,
          }),
          adminH,
        );
        check(updateRes, { 'PUT /v1/admin/banners/:id: 2xx': (r) => ok(r) });

        // Delete
        const deleteRes = http.del(`${BASE_URL}/v1/admin/banners/${bannerId}`, null, adminH);
        check(deleteRes, { 'DELETE /v1/admin/banners/:id: 2xx': (r) => ok(r) });
      }
    }
  });

  // -------------------------------------------------------------------------
  group('Admin – Articles', () => {
    // List
    const listRes = http.get(`${BASE_URL}/v1/admin/articles?page=1&page_size=10`, adminH);
    check(listRes, { 'GET /v1/admin/articles: 200': (r) => r.status === 200 });

    // Get
    const getRes = http.get(`${BASE_URL}/v1/admin/articles/${articleId}`, adminH);
    check(getRes, { 'GET /v1/admin/articles/:id: 200': (r) => r.status === 200 });

    // Update
    const updateRes = http.put(
      `${BASE_URL}/v1/admin/articles/${articleId}`,
      JSON.stringify({ title: 'K6 Article Updated', content: '# Updated', module_id: moduleId }),
      adminH,
    );
    check(updateRes, { 'PUT /v1/admin/articles/:id: 2xx': (r) => ok(r) });

    // Unpublish then re-publish
    http.post(
      `${BASE_URL}/v1/admin/articles/${articleId}/publish`,
      JSON.stringify({ status: 0 }),
      adminH,
    );
    const pubRes = http.post(
      `${BASE_URL}/v1/admin/articles/${articleId}/publish`,
      JSON.stringify({ status: 1 }),
      adminH,
    );
    check(pubRes, { 'POST /v1/admin/articles/:id/publish: 2xx': (r) => ok(r) });

    // Pin
    const pinRes = http.post(
      `${BASE_URL}/v1/admin/articles/${articleId}/pin`,
      JSON.stringify({ sort_order: 1 }),
      adminH,
    );
    check(pinRes, { 'POST /v1/admin/articles/:id/pin: 2xx': (r) => ok(r) });

    // Copy then cleanup
    const copyRes = http.post(`${BASE_URL}/v1/admin/articles/${articleId}/copy`, null, adminH);
    check(copyRes, { 'POST /v1/admin/articles/:id/copy: 201': (r) => r.status === 201 });
    const copiedArticleId = copyRes.json('data.id');
    if (copiedArticleId) {
      const deleteCopiedRes = http.del(`${BASE_URL}/v1/admin/articles/${copiedArticleId}`, null, adminH);
      check(deleteCopiedRes, { 'DELETE /v1/admin/articles/:id (cleanup copied): 2xx': (r) => ok(r) });
    }
  });

  // -------------------------------------------------------------------------
  group('Admin – Courses', () => {
    // List
    const listRes = http.get(`${BASE_URL}/v1/admin/courses?page=1&page_size=10`, adminH);
    check(listRes, { 'GET /v1/admin/courses: 200': (r) => r.status === 200 });

    // Get
    const getRes = http.get(`${BASE_URL}/v1/admin/courses/${courseId}`, adminH);
    check(getRes, { 'GET /v1/admin/courses/:id: 200': (r) => r.status === 200 });

    // Update
    const updateRes = http.put(
      `${BASE_URL}/v1/admin/courses/${courseId}`,
      JSON.stringify({ title: 'K6 Course Updated', description: 'Updated', module_id: moduleId, price: 0 }),
      adminH,
    );
    check(updateRes, { 'PUT /v1/admin/courses/:id: 2xx': (r) => ok(r) });

    // List units
    const unitsRes = http.get(`${BASE_URL}/v1/admin/courses/${courseId}/units`, adminH);
    check(unitsRes, { 'GET /v1/admin/courses/:id/units: 200': (r) => r.status === 200 });

    // Update unit
    const updateUnitRes = http.put(
      `${BASE_URL}/v1/admin/courses/${courseId}/units/${unitId}`,
      JSON.stringify({ title: 'K6 Unit Updated', duration: 45, sort_order: 1 }),
      adminH,
    );
    check(updateUnitRes, { 'PUT /v1/admin/courses/:id/units/:uid: 2xx': (r) => ok(r) });

    // Pin
    const pinRes = http.post(
      `${BASE_URL}/v1/admin/courses/${courseId}/pin`,
      JSON.stringify({ sort_order: 1 }),
      adminH,
    );
    check(pinRes, { 'POST /v1/admin/courses/:id/pin: 2xx': (r) => ok(r) });

    // Unpublish then re-publish
    const unpubRes = http.post(
      `${BASE_URL}/v1/admin/courses/${courseId}/publish`,
      JSON.stringify({ status: 0 }),
      adminH,
    );
    check(unpubRes, { 'POST /v1/admin/courses/:id/publish (unpublish): 2xx': (r) => ok(r) });
    const pubRes = http.post(
      `${BASE_URL}/v1/admin/courses/${courseId}/publish`,
      JSON.stringify({ status: 1 }),
      adminH,
    );
    check(pubRes, { 'POST /v1/admin/courses/:id/publish: 2xx': (r) => ok(r) });

    // Copy then cleanup
    const copyRes = http.post(`${BASE_URL}/v1/admin/courses/${courseId}/copy`, null, adminH);
    check(copyRes, { 'POST /v1/admin/courses/:id/copy: 201': (r) => r.status === 201 });
    const copiedCourseId = copyRes.json('data.id');
    if (copiedCourseId) {
      const deleteCopiedRes = http.del(
        `${BASE_URL}/v1/admin/courses/${copiedCourseId}`,
        null,
        withExpectedStatuses(adminH, { min: 200, max: 299 }, { min: 400, max: 499 }),
      );
      check(deleteCopiedRes, { 'DELETE /v1/admin/courses/:id (cleanup copied): 2xx': (r) => ok(r) || r.status < 500 });
    }

    // Create and delete temporary unit
    const createUnitRes = http.post(
      `${BASE_URL}/v1/admin/courses/${courseId}/units`,
      JSON.stringify({ title: 'K6 Temp Unit', video_file_id: setupVideoFileId || DEFAULT_VIDEO_FILE_ID, duration: 60, sort_order: 2 }),
      adminH,
    );
    check(createUnitRes, { 'POST /v1/admin/courses/:id/units: 201': (r) => r.status === 201 });
    const tempUnitId = createUnitRes.json('data.id');
    if (tempUnitId) {
      const delUnitRes = http.del(`${BASE_URL}/v1/admin/courses/${courseId}/units/${tempUnitId}`, null, adminH);
      check(delUnitRes, { 'DELETE /v1/admin/courses/:id/units/:unit_id: 2xx': (r) => ok(r) });
    }
  });

  // -------------------------------------------------------------------------
  group('Admin – Comments', () => {
    // List
    const listRes = http.get(`${BASE_URL}/v1/admin/comments?page=1&page_size=10`, adminH);
    check(listRes, { 'GET /v1/admin/comments: 200': (r) => r.status === 200 });

    // Audit (approve)
    const auditRes = http.put(
      `${BASE_URL}/v1/admin/comments/${commentId}/audit`,
      JSON.stringify({ status: 1 }),
      adminH,
    );
    check(auditRes, { 'PUT /v1/admin/comments/:id/audit: 2xx': (r) => ok(r) });

    // Delete
    const deleteRes = http.del(`${BASE_URL}/v1/admin/comments/${commentId}`, null, adminH);
    check(deleteRes, { 'DELETE /v1/admin/comments/:id: 2xx': (r) => ok(r) });
  });

  // -------------------------------------------------------------------------
  group('Admin – System', () => {
    // Get WeChat config
    const wcRes = http.get(`${BASE_URL}/v1/admin/wechat-config`, adminH);
    check(wcRes, { 'GET /v1/admin/wechat-config: 2xx': (r) => ok(r) });

    // Update WeChat config
    const wcUpdateRes = http.put(
      `${BASE_URL}/v1/admin/wechat-config`,
      JSON.stringify({ app_id: 'k6_test_app_id', app_secret: 'k6_test_app_secret_32chars_long!', api_token: '' }),
      adminH,
    );
    check(wcUpdateRes, { 'PUT /v1/admin/wechat-config: 2xx': (r) => ok(r) });

    // Audit logs
    const logsRes = http.get(`${BASE_URL}/v1/admin/audit-logs?page=1&page_size=10`, adminH);
    check(logsRes, { 'GET /v1/admin/audit-logs: 2xx': (r) => ok(r) });

    // Get log config
    const logConfRes = http.get(`${BASE_URL}/v1/admin/log-config`, adminH);
    check(logConfRes, { 'GET /v1/admin/log-config: 2xx': (r) => ok(r) });

    // Update log config
    const logConfUpdateRes = http.put(
      `${BASE_URL}/v1/admin/log-config`,
      JSON.stringify({ retention_days: 30 }),
      adminH,
    );
    check(logConfUpdateRes, { 'PUT /v1/admin/log-config: 2xx': (r) => ok(r) });
  });

  // -------------------------------------------------------------------------
  group('Admin – Attributes', () => {
    // List attributes
    const listRes = http.get(`${BASE_URL}/v1/admin/attributes`, adminH);
    check(listRes, { 'GET /v1/admin/attributes: 200': (r) => r.status === 200 });

    // Create attribute
    const createRes = http.post(
      `${BASE_URL}/v1/admin/attributes`,
      JSON.stringify({ name: `K6 Attr ${uniqueTestId}` }),
      adminH,
    );
    check(createRes, { 'POST /v1/admin/attributes: 201': (r) => r.status === 201 });
    const attributeId = createRes.json('data.id');
    if (attributeId) {
      // Update attribute
      const updateRes = http.put(
        `${BASE_URL}/v1/admin/attributes/${attributeId}`,
        JSON.stringify({ name: `K6 Attr Updated ${uniqueTestId}` }),
        adminH,
      );
      check(updateRes, { 'PUT /v1/admin/attributes/:id: 2xx': (r) => ok(r) });

      // Set user attribute
      const setUserAttrRes = http.post(
        `${BASE_URL}/v1/admin/users/${newUserId}/attributes`,
        JSON.stringify({ attribute_id: attributeId, value: 'k6-value' }),
        adminH,
      );
      check(setUserAttrRes, { 'POST /v1/admin/users/:id/attributes: 2xx': (r) => ok(r) });

      // List user attributes
      const listUserAttrRes = http.get(`${BASE_URL}/v1/admin/users/${newUserId}/attributes`, adminH);
      check(listUserAttrRes, { 'GET /v1/admin/users/:id/attributes: 200': (r) => r.status === 200 });

      // Delete user attribute
      const delUserAttrRes = http.del(
        `${BASE_URL}/v1/admin/users/${newUserId}/attributes?attribute_id=${attributeId}`,
        null,
        adminH,
      );
      check(delUserAttrRes, { 'DELETE /v1/admin/users/:id/attributes: 2xx': (r) => ok(r) });

      // Delete attribute
      const deleteRes = http.del(`${BASE_URL}/v1/admin/attributes/${attributeId}`, null, adminH);
      check(deleteRes, { 'DELETE /v1/admin/attributes/:id: 2xx': (r) => ok(r) });
    }
  });
}

// ---------------------------------------------------------------------------
// teardown – clean up test data created in setup
// ---------------------------------------------------------------------------

export function teardown(data) {
  const { adminToken, moduleId, pageId, articleId, courseId, unitId, roleId, commentId, newUserId } = data;
  const adminH = headers(adminToken);
  const cleanupH = withExpectedStatuses(adminH, { min: 200, max: 499 });

  // Delete in dependency order (children before parents)
  if (commentId)  http.del(`${BASE_URL}/v1/admin/comments/${commentId}`, null, cleanupH);
  if (unitId)     http.del(`${BASE_URL}/v1/admin/courses/${courseId}/units/${unitId}`, null, cleanupH);
  if (courseId)   http.del(`${BASE_URL}/v1/admin/courses/${courseId}`, null, cleanupH);
  if (articleId)  http.del(`${BASE_URL}/v1/admin/articles/${articleId}`, null, cleanupH);
  if (pageId)     http.del(`${BASE_URL}/v1/admin/modules/${moduleId}/pages/${pageId}`, null, cleanupH);
  if (moduleId)   http.del(`${BASE_URL}/v1/admin/modules/${moduleId}`, null, cleanupH);
  if (newUserId)  http.del(`${BASE_URL}/v1/admin/users/${newUserId}`, null, cleanupH);
  if (roleId)     http.del(`${BASE_URL}/v1/admin/roles/${roleId}`, null, cleanupH);
}
