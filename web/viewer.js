// Three.js-backed 3MF viewer. Hydrates every <div class="model-viewer"
// data-model="path/to.3mf"> on the page: fits the camera to the model
// bounding box, adds ambient + directional lighting, and wires up
// OrbitControls so the user can spin / zoom the part.
//
// Loaded as a module from <script type="module" src="…/js/viewer.js">.
// Pages must include an importmap pointing "three" and "three/addons/"
// at a CDN copy of three.js (≥ r150).

import * as THREE from 'three';
import { ThreeMFLoader } from 'three/addons/loaders/3MFLoader.js';
import { OrbitControls } from 'three/addons/controls/OrbitControls.js';

const loader = new ThreeMFLoader();

document.querySelectorAll('div.model-viewer[data-model]').forEach(initViewer);

function initViewer(div) {
  const src = div.dataset.model;
  const width = () => div.clientWidth;
  const height = () => div.clientHeight || 400;

  const scene = new THREE.Scene();
  scene.background = new THREE.Color(0xfafafa);

  const camera = new THREE.PerspectiveCamera(40, width() / height(), 0.1, 2000);
  camera.position.set(60, 40, 60);

  const renderer = new THREE.WebGLRenderer({ antialias: true });
  renderer.setSize(width(), height());
  renderer.setPixelRatio(window.devicePixelRatio);
  div.appendChild(renderer.domElement);

  scene.add(new THREE.AmbientLight(0xffffff, 0.55));
  const sun = new THREE.DirectionalLight(0xffffff, 0.85);
  sun.position.set(1, 1.2, 0.8);
  scene.add(sun);
  const fill = new THREE.DirectionalLight(0xffffff, 0.25);
  fill.position.set(-0.8, -0.4, -0.6);
  scene.add(fill);

  const controls = new OrbitControls(camera, renderer.domElement);
  controls.enableDamping = true;
  controls.dampingFactor = 0.08;

  loader.load(
    src,
    (object) => {
      // Centre on origin and frame the camera to fit.
      const box = new THREE.Box3().setFromObject(object);
      const size = box.getSize(new THREE.Vector3());
      const centre = box.getCenter(new THREE.Vector3());
      object.position.sub(centre);

      const maxDim = Math.max(size.x, size.y, size.z);
      const dist = maxDim / Math.tan((Math.PI * camera.fov) / 360) * 0.9;
      camera.position.set(dist * 0.7, dist * 0.55, dist * 0.7);
      camera.near = dist / 100;
      camera.far = dist * 10;
      camera.updateProjectionMatrix();
      controls.target.set(0, 0, 0);
      controls.update();

      scene.add(object);
    },
    undefined,
    (err) => {
      console.error(`viewer: failed to load ${src}`, err);
      div.textContent = `Failed to load ${src}`;
    },
  );

  function animate() {
    requestAnimationFrame(animate);
    controls.update();
    renderer.render(scene, camera);
  }
  animate();

  window.addEventListener('resize', () => {
    camera.aspect = width() / height();
    camera.updateProjectionMatrix();
    renderer.setSize(width(), height());
  });
}
