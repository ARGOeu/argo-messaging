// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'ARGO Messaging Documentation',
  tagline: 'Learn how the ARGO Messaging Service (AMS) works',
  url: 'https://argoeu.github.io',
  baseUrl: '/argo-messaging/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/favicon.ico',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'ARGOeu', // Usually your GitHub org/user name.
  projectName: 'argo-messaging', // Usually your repo name.

  // Even if you don't use internalization, you can use this field to set useful
  // metadata like html lang. For example, if your site is Chinese, you may want
  // to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          // editUrl:
          //   'https://github.com/facebook/docusaurus/tree/main/packages/create-docusaurus/templates/shared/',
        },
        blog: false,
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      navbar: {
        title: 'ARGO Messaging Service',
        logo: {
          alt: 'argo-messaging logo',
          src: 'img/ams-logo.svg',
        },
        items: [
          {
            to: 'docs/',
            activeBasePath: 'docs',
            label: 'Docs',
            position: 'left',
          },
          {
            href: 'pathname:///openapi/explore',
            label: 'Explore the API',
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
        logo: {
          alt: 'GRNET',
          src: 'img/grnet-logo.png',
          href: 'http://www.grnet.gr/',
        },
        links: [
          {
            title: 'Docs',
            items: [
              {
                label: 'Explore Documentation',
                to: 'docs/',
              },
              {
                to: 'pathname:///openapi/explore',
                label: 'Explore the API (openapi v3)',
                position: 'left',
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
        copyright: `Copyright © ${new Date().getFullYear()} <a href="http://www.grnet.gr/"> GRNET </a>`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
      },
    }),
    themes: [
      [
        "@easyops-cn/docusaurus-search-local",
        {
          hashed: true,
          language: ["en", "zh"],
          highlightSearchTermsOnTargetPage: true,
          explicitSearchResultPath: true,
          indexBlog: false,
        },
      ],
    ],
};

module.exports = config;
