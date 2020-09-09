module.exports = {
  title: 'ARGO Messaging Documentation',
  tagline: 'Learn how the ARGO Messaging Service (AMS) works',
  url: 'https://argoeu.github.io',
  baseUrl: '/argo-messaging/',
  onBrokenLinks: 'throw',
  favicon: 'img/favicon.ico',
  organizationName: 'ARGOeu', // Usually your GitHub org/user name.
  projectName: 'argo-messaging', // Usually your repo name.
  themeConfig: {
    navbar: {
      title: 'ARGO Messaging Service',
      logo: {
        alt: 'argo-messaging logo',
        src: 'img/ams.png',
      },
      items: [
        {
          to: 'docs/',
          activeBasePath: 'docs',
          label: 'Docs',
          position: 'left',
        },
        {
          href: 'https://github.com/ARGOeu/argo-messaging',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Docs',
          items: [
            {
              label: 'Explore Documentation',
              to: 'docs/',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'Github',
              href: 'https://github.com/ARGOeu/argo-messaging',
            }
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/ARGOeu/argo-messaging',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} GRNET`,
    },
  },
  presets: [
    [
      '@docusaurus/preset-classic',
      {
        docs: {
          // It is recommended to set document id as docs home page (`docs/` path).
          homePageId: 'overview',
          sidebarPath: require.resolve('./sidebars.js'),
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      },
    ],
  ],
};
