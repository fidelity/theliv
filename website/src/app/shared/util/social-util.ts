export class SocialUtil {
    isThumbsUp(event: any, resource: any): void {
        resource.thumbs = !resource.thumbs;
        event.stopPropagation();
    }

    openShared(event: any, resource: any): void {
        resource.shared++;
        event.stopPropagation();
    }

    copyYamlCode(code: any): void {
        const selBox = document.createElement('textarea');
        selBox.style.position = 'fixed';
        selBox.style.left = '0';
        selBox.style.top = '0';
        selBox.style.opacity = '0';
        selBox.value = code;
        document.body.appendChild(selBox);
        selBox.focus();
        selBox.select();
        document.execCommand('copy');
        document.body.removeChild(selBox);
    }
}
