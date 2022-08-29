import React from 'react';
import clsx from 'clsx';
import styles from './styles.module.css';

const FeatureList = [
  {
    title: 'Create Topics',
    Svg: require('@site/static/img/undraw_topics.svg').default,
    description: (
      <>
        Create Topics as destinations to send messages to
      </>
    ),
  },
  {
    title: 'Create Subscriptions',
    Svg: require('@site/static/img/undraw_subs.svg').default,
    description: (
      <>
        Create Subscriptions on topics to receive messages. 
      </>
    ),
  },
  {
    title: 'Send and receive messages',
    Svg: require('@site/static/img/undraw_msg.svg').default,
    description: (
      <>
        Use simple HTTP calls to publish and receive messages easily
      </>
    ),
  },
];

function Feature({Svg, title, description}) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center">
        <Svg className={styles.featureSvg} role="img" />
      </div>
      <div className="text--center padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
