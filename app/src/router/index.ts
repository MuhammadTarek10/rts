import { createRouter, createWebHistory, RouterView } from 'vue-router';
import HomePage from '../pages/HomePage.vue';
import { useUserStore } from '../stores/user';

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomePage,
      meta: {
        layout: 'MainLayout',
        requiresAuth: true,
      },
    },
    {
      path: '/about',
      name: 'about',
      component: () => import('../pages/AboutPage.vue'),
      meta: {
        layout: 'MainLayout',
        requiresAuth: true,
      },
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
          component: () => import('../pages/auth/SignInPage.vue'),
        },
        {
          path: 'sign-up',
          name: 'sign-up',
          component: () => import('../pages/auth/SignUpPage.vue'),
        },
        {
          path: 'forgot-password',
          name: 'forgot-password',
          component: () => import('../pages/auth/ForgotPasswordPage.vue'),
        },
      ],
    },
  ],
});

router.beforeEach((to, _from, next) => {
  const isProtected = to.matched.some(
    (record) => record.meta.requiresAuth === true
  );
  const isAuthRoute = to.matched.some((record) =>
    record.path.startsWith('/auth')
  );
  const isAuthenticated = useUserStore().user !== null;

  if (isProtected && !isAuthenticated) {
    next({ name: 'sign-in' });
  } else if (isAuthRoute && isAuthenticated) {
    next({ name: 'home' });
  } else {
    next();
  }
});

export default router;
