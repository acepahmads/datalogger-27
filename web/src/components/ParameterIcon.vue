<template>
  <span
    class="inline-flex items-center justify-center flex-shrink-0 transition-colors"
    :class="[
      sizeClass,
      showBg ? iconDef.bgClass + ' p-0.5 rounded-md' : '',
      customColor ? customColor : iconDef.colorClass
    ]"
    :title="title || iconDef.label"
    v-html="iconDef.svg"
  ></span>
</template>

<script>
import { resolveParameterIconDef } from '../utils/parameterIcons.js';

export default {
  name: 'ParameterIcon',
  props: {
    icon: {
      type: String,
      default: '',
    },
    code: {
      type: String,
      default: '',
    },
    name: {
      type: String,
      default: '',
    },
    unit: {
      type: String,
      default: '',
    },
    sizeClass: {
      type: String,
      default: 'w-4 h-4',
    },
    customColor: {
      type: String,
      default: '',
    },
    showBg: {
      type: Boolean,
      default: false,
    },
    title: {
      type: String,
      default: '',
    },
  },
  computed: {
    iconDef() {
      return resolveParameterIconDef(this.icon, this.code, this.name, this.unit);
    },
  },
};
</script>
