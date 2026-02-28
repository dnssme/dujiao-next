<script setup lang="ts">
import { computed } from 'vue'
import { storeToRefs } from 'pinia'
import { useConfirmStore } from '@/stores/confirm'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'

const confirmStore = useConfirmStore()
const { open, options } = storeToRefs(confirmStore)

const confirmVariant = computed(() => (options.value.variant === 'destructive' ? 'destructive' : 'default'))
</script>

<template>
  <Dialog :open="open" @update:open="confirmStore.setOpen">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>{{ options.title }}</DialogTitle>
        <DialogDescription>
          {{ options.description }}
        </DialogDescription>
      </DialogHeader>
      <DialogFooter>
        <Button variant="outline" @click="confirmStore.cancel">{{ options.cancelText }}</Button>
        <Button :variant="confirmVariant" @click="confirmStore.confirm">{{ options.confirmText }}</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
