import { createRouter, createWebHistory, RouterView } from 'vue-router';
import HomePage from '@/pages/HomePage.vue';
import { useUserStore } from '@/stores/user';

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomePage,
      meta: { layout: 'MainLayout', requiresAuth: true },
    },
    {
      path: '/about',
      name: 'about',
      component: () => import('@/pages/AboutPage.vue'),
      meta: { layout: 'MainLayout', requiresAuth: true },
    },
    {
      path: '/profile',
      name: 'profile',
      component: () => import('@/pages/ProfilePage.vue'),
      meta: { layout: 'MainLayout', requiresAuth: true },
    },
    {
      path: '/auth',
      name: 'auth',
      component: RouterView,
      meta: { layout: 'AuthLayout' },
      children: [
        {
          path: 'sign-in',
          name: 'sign-in',
          component: () => import('@/pages/auth/SignInPage.vue'),
        },
        {
          path: 'sign-up',
          name: 'sign-up',
          component: () => import('@/pages/auth/SignUpPage.vue'),
        },
        {
          path: 'forgot-password',
          name: 'forgot-password',
          component: () => import('@/pages/auth/ForgotPasswordPage.vue'),
        },
      ],
    },
  ],
});

router.beforeEach(async (to) => {
  const userStore = useUserStore();

  if (!userStore.isInitialized) {
    await userStore.fetchUser();
  }

  const isProtected = to.matched.some((r) => r.meta.requiresAuth === true);
  const isAuthRoute = to.matched.some((r) => r.path.startsWith('/auth'));

  if (isProtected && !userStore.isAuthenticated) {
    return { name: 'sign-in', query: { redirect: to.fullPath } };
  }

  if (isAuthRoute && userStore.isAuthenticated) {
    return { name: 'home' };
  }
});

export default router;
