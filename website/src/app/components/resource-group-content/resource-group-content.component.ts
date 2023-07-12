/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
import { Component, OnInit } from '@angular/core';
import { faExclamationTriangle, faLightbulb, faTimes, faThumbsUp, faExternalLinkAlt, 
  faClipboard, faShareAlt, faCopy, faDownload, faBookReader, faArrowLeft, faEdit } from '@fortawesome/free-solid-svg-icons';
import { ActivatedRoute } from '@angular/router';
import { KubernetesService } from 'src/app/services/kubernetes.service';
import { SocialUtil } from '../../shared/util/social-util';
const yaml = require('js-yaml');

@Component({
  selector: 'app-resource-group-content',
  templateUrl: './resource-group-content.component.html',
  styleUrls: ['./resource-group-content.component.scss']
})
export class ResourceGroupContentComponent implements OnInit {
  resourceGroup: any;
  selectedResource: any;
  containerResource: any;
  podResource: any;
  replicasetResource: any;
  deployResource: any;
  svcResource: any;
  ingressResource: any;
  otherResource: any;
  gridToggle = false;
  popupResource: any;
  popupTemplate: any;
  
  faExclamationTriangle = faExclamationTriangle;
  faLightbulb = faLightbulb;
  faEdit = faEdit;
  faTimes = faTimes;
  faExternalLinkAlt=faExternalLinkAlt
  faThumbsUp = faThumbsUp;
  faShareAlt = faShareAlt;
  faCopy = faCopy;
  faDownload = faDownload;
  faClipboard = faClipboard;
  faBookReader = faBookReader;
  faArrowLeft = faArrowLeft;

  ns: any;
  cluster: any;

  constructor(
    private kubeService: KubernetesService,
    public socialUtil: SocialUtil,
    private route: ActivatedRoute
  ) { }

  ngOnInit(): void {
    this.resourceGroup = this.kubeService.resourceList$.getValue();
    this.route.params.subscribe(params => {
      this.selectedResource = this.resourceGroup.find((item: any) => item.id === params.id);
      this.containerResource = this.selectedResource.resources.filter((item: any) => item.type === 'container');
      this.podResource = this.selectedResource.resources.filter((item: any) => item.type === 'pod');
      this.replicasetResource = this.selectedResource.resources.filter((item: any) => item.type === 'replicaset');
      this.deployResource = this.selectedResource.resources.filter((item: any) => item.type === 'deployment');
      this.svcResource = this.selectedResource.resources.filter((item: any) => item.type === 'service');
      this.ingressResource = this.selectedResource.resources.filter((item: any) => item.type === 'ingress');
      this.otherResource = this.selectedResource.resources.filter((item: any) => item.type !== 'pod' && item.type !== 'deployment' && item.type !== 'service' && item.type !== 'ingress' && item.type !== 'container' && item.type !== 'replicaset');
    });
    this.ns = this.kubeService.selectedNs$.getValue();
    this.cluster = this.kubeService.selectedClusters$.getValue();
  }

  showPopupItem(resource: any): void {
    this.popupTemplate = yaml.safeDump(resource.metadata);
    this.popupResource = resource.name;
    this.gridToggle = true;
  }

  close(): void {
    this.gridToggle = false;
  }
}
