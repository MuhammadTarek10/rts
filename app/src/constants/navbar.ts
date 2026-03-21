interface NavbarLink {
  title: string;
  link: string;
}

export const NAVBAR_PAGES: NavbarLink[] = [
  {
    title: 'Home',
    link: '/',
  },
  {
    title: 'About',
    link: '/about',
  },
  {
    title: 'Status',
    link: '/status',
  },
];
